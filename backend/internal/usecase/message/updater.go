package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
    "github.com/newt239/chat/internal/domain/service"
    domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

// MessageUpdater はメッセージ更新を担当するユースケースです
type MessageUpdater struct {
	messageRepo           domainrepository.MessageRepository
	channelRepo           domainrepository.ChannelRepository
	channelMemberRepo     domainrepository.ChannelMemberRepository
	workspaceRepo         domainrepository.WorkspaceRepository
	userRepo              domainrepository.UserRepository
	userGroupRepo         domainrepository.UserGroupRepository
	userMentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo      domainrepository.MessageGroupMentionRepository
	linkRepo              domainrepository.MessageLinkRepository
	attachmentRepo        domainrepository.AttachmentRepository
	notificationSvc       service.NotificationService
	mentionService        service.MentionService
	linkProcessingService service.LinkProcessingService
	transactionManager    transaction.Manager
	assembler             *MessageOutputAssembler
    channelAccessSvc      domainservice.ChannelAccessService
}

// NewMessageUpdater は新しいMessageUpdaterを作成します
func NewMessageUpdater(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	notificationSvc service.NotificationService,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
    channelAccessSvc domainservice.ChannelAccessService,
) *MessageUpdater {
	return &MessageUpdater{
		messageRepo:           messageRepo,
		channelRepo:           channelRepo,
		channelMemberRepo:     channelMemberRepo,
		workspaceRepo:         workspaceRepo,
		userRepo:              userRepo,
		userGroupRepo:         userGroupRepo,
		userMentionRepo:       userMentionRepo,
		groupMentionRepo:      groupMentionRepo,
		linkRepo:              linkRepo,
		attachmentRepo:        attachmentRepo,
		notificationSvc:       notificationSvc,
		mentionService:        mentionService,
		linkProcessingService: linkProcessingService,
		transactionManager:    transactionManager,
		assembler:             NewMessageOutputAssembler(),
        channelAccessSvc:      channelAccessSvc,
	}
}

// UpdateMessage はメッセージを更新します
func (u *MessageUpdater) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error) {
	// メッセージ存在確認
	message, err := u.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("メッセージの取得に失敗しました: %w", err)
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// チャンネルアクセス確認
    channel, err := u.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.EditorID)
	if err != nil {
		return nil, err
	}

	// 削除済みメッセージの編集禁止
	if message.DeletedAt != nil {
		return nil, ErrCannotEditDeleted
	}

	// 権限確認: 投稿者本人または管理者
	canEdit, err := u.canModifyMessage(ctx, channel.WorkspaceID, message.UserID, input.EditorID)
	if err != nil {
		return nil, fmt.Errorf("権限確認に失敗しました: %w", err)
	}
	if !canEdit {
		return nil, ErrUnauthorized
	}

	var result *MessageOutput
	err = u.transactionManager.Do(ctx, func(txCtx context.Context) error {
		// メッセージ本文を更新
		message.Body = input.Body
		now := time.Now()
		message.EditedAt = &now

		// データベース更新
		if err := u.messageRepo.Update(txCtx, message); err != nil {
			return fmt.Errorf("メッセージの更新に失敗しました: %w", err)
		}

		// 既存のメンション・リンクを削除
		if err := u.userMentionRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete user mentions: %w", err)
		}
		if err := u.groupMentionRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete group mentions: %w", err)
		}
		if err := u.linkRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete links: %w", err)
		}

		// 新しいメンション・リンクを抽出・保存
		if err := u.extractAndSaveMentionsAndLinks(txCtx, message.ID, input.Body, channel.WorkspaceID); err != nil {
			return fmt.Errorf("failed to extract and save mentions/links: %w", err)
		}

		// 更新後のデータを取得してMessageOutputを構築
		userMentions, err := u.userMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user mentions: %w", err)
		}

		groupMentions, err := u.groupMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch group mentions: %w", err)
		}

		links, err := u.linkRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch links: %w", err)
		}

		reactions, err := u.messageRepo.FindReactions(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch reactions: %w", err)
		}

		attachmentList, err := u.attachmentRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch attachments: %w", err)
		}

		// ユーザー情報を取得
		user, err := u.userRepo.FindByID(txCtx, message.UserID)
		if err != nil {
			return fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
		}

		// グループ情報を取得
		groupIDs := make([]string, 0)
		groupIDSet := make(map[string]bool)
		for _, gm := range groupMentions {
			if !groupIDSet[gm.GroupID] {
				groupIDs = append(groupIDs, gm.GroupID)
				groupIDSet[gm.GroupID] = true
			}
		}

		groups := make(map[string]*entity.UserGroup)
		if len(groupIDs) > 0 {
			groupList, err := u.userGroupRepo.FindByIDs(txCtx, groupIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch groups: %w", err)
			}
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}

		userMap := map[string]*entity.User{user.ID: user}
		output := u.assembler.AssembleMessageOutput(message, user, userMentions, groupMentions, links, reactions, attachmentList, groups, userMap)
		result = &output

		return nil
	})

	if err != nil {
		return nil, err
	}

	// WebSocket通知を送信
	if u.notificationSvc != nil {
		u.notificationSvc.NotifyUpdatedMessage(channel.WorkspaceID, channel.ID, *result)
	}

	return result, nil
}

// ensureChannelAccess は ChannelAccessService に委譲済み

// canModifyMessage はユーザーがメッセージを編集・削除できるかどうかを確認します
func (u *MessageUpdater) canModifyMessage(ctx context.Context, workspaceID, messageOwnerID, executorID string) (bool, error) {
	// 投稿者本人の場合は許可
	if messageOwnerID == executorID {
		return true, nil
	}

	// 管理者権限チェック
	member, err := u.workspaceRepo.FindMember(ctx, workspaceID, executorID)
	if err != nil {
		return false, fmt.Errorf("ワークスペースメンバー情報の取得に失敗しました: %w", err)
	}
	if member == nil {
		return false, nil
	}

	// owner または admin の場合は許可
	if member.Role == entity.WorkspaceRoleOwner || member.Role == entity.WorkspaceRoleAdmin {
		return true, nil
	}

	return false, nil
}

// extractAndSaveMentionsAndLinks はメンションとリンクの抽出・保存を行います
func (u *MessageUpdater) extractAndSaveMentionsAndLinks(ctx context.Context, messageID, body, workspaceID string) error {
	// ユーザーメンションの抽出
	userMentions, err := u.mentionService.ExtractUserMentions(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to extract user mentions: %w", err)
	}
	for _, mention := range userMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := u.userMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}

	// グループメンションの抽出
	groupMentions, err := u.mentionService.ExtractGroupMentions(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to extract group mentions: %w", err)
	}
	for _, mention := range groupMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := u.groupMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	// リンクの抽出とOGP取得
	links, err := u.linkProcessingService.ProcessLinks(ctx, body)
	if err != nil {
		return fmt.Errorf("failed to process links: %w", err)
	}

	for _, link := range links {
		// 既存のリンクをチェック
		existingLink, err := u.linkRepo.FindByURL(ctx, link.URL)
		if err != nil {
			continue // エラーは無視
		}

		if existingLink != nil {
			// 既存のリンクを再利用
			link.MessageID = messageID
			link.Title = existingLink.Title
			link.Description = existingLink.Description
			link.ImageURL = existingLink.ImageURL
			link.SiteName = existingLink.SiteName
			link.CardType = existingLink.CardType
			link.CreatedAt = time.Now()
		} else {
			// 新しいリンクを保存
			link.MessageID = messageID
			link.CreatedAt = time.Now()
		}

		// リンクを保存
		if err := u.linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}

	return nil
}
