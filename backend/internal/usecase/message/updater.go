package message

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
)

// MessageUpdater はメッセージ更新を担当するユースケースです
type MessageUpdater struct {
	messageRepo        domainrepository.MessageRepository
	userRepo           domainrepository.UserRepository
	workspaceRepo      domainrepository.WorkspaceRepository
	notificationSvc    Notifier
	recorder           *contentRecorder
	transactionManager transaction.Manager
	outputBuilder      *MessageOutputBuilder
	channelAccessSvc   service.ChannelAccessService
	searchIndexer      SearchIndexer
}

// NewMessageUpdater は新しいMessageUpdaterを作成します
func NewMessageUpdater(
	messageRepo domainrepository.MessageRepository,
	userRepo domainrepository.UserRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	notificationSvc Notifier,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
	outputBuilder *MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
	searchIndexer SearchIndexer,
) *MessageUpdater {
	return &MessageUpdater{
		messageRepo:     messageRepo,
		userRepo:        userRepo,
		workspaceRepo:   workspaceRepo,
		notificationSvc: notificationSvc,
		recorder: &contentRecorder{
			mentionService:        mentionService,
			userMentionRepo:       userMentionRepo,
			groupMentionRepo:      groupMentionRepo,
			linkProcessingService: linkProcessingService,
			linkRepo:              linkRepo,
		},
		transactionManager: transactionManager,
		outputBuilder:      outputBuilder,
		channelAccessSvc:   channelAccessSvc,
		searchIndexer:      searchIndexer,
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

	if err := ensureNotOfficial(ctx, u.userRepo, message); err != nil {
		return nil, err
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
		applyBroadcastMentions(message)
		now := time.Now()
		message.EditedAt = &now

		// データベース更新
		if err := u.messageRepo.Update(txCtx, message); err != nil {
			return fmt.Errorf("メッセージの更新に失敗しました: %w", err)
		}

		// 既存のメンション・リンクを入れ替える
		previous, err := u.recorder.userMentionRepo.FindByMessageIDs(txCtx, []string{message.ID})
		if err != nil {
			return fmt.Errorf("failed to load user mentions: %w", err)
		}
		if err := u.recorder.userMentionRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete user mentions: %w", err)
		}
		if err := u.recorder.groupMentionRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete group mentions: %w", err)
		}
		if err := u.recorder.linkRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete links: %w", err)
		}
		if err := u.recorder.record(txCtx, message.ID, input.Body, channel.WorkspaceID, previous); err != nil {
			return err
		}

		outputs, err := u.outputBuilder.Build(txCtx, input.EditorID, []*entity.Message{message})
		if err != nil {
			return err
		}
		result = &outputs[0]
		return nil
	})
	if err != nil {
		return nil, err
	}

	u.searchIndexer.Sync(ctx, message.ID)

	// WebSocket通知を送信
	if u.notificationSvc != nil {
		u.notificationSvc.NotifyUpdatedMessage(channel.WorkspaceID, channel.ID, result.ForBroadcast())
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
