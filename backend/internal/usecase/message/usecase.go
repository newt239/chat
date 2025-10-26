package message

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/infrastructure/logger"
	"go.uber.org/zap"
)

var (
	ErrChannelNotFound       = errors.New("channel not found")
	ErrUnauthorized          = errors.New("unauthorized to perform this action")
	ErrParentMessageNotFound = errors.New("parent message not found")
	ErrMessageNotFound       = errors.New("message not found")
	ErrMessageAlreadyDeleted = errors.New("message already deleted")
	ErrCannotEditDeleted     = errors.New("cannot edit deleted message")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

// RelatedData はメッセージに関連するデータをまとめた構造体です
type RelatedData struct {
	UserMentions  []*entity.MessageUserMention
	GroupMentions []*entity.MessageGroupMention
	Links         []*entity.MessageLink
	Reactions     map[string][]*entity.MessageReaction
	Attachments   map[string][]*entity.Attachment
}

// MessageOutputAssembler はMessageOutputの構築を担当するコンポーネントです
type MessageOutputAssembler struct{}

// NewMessageOutputAssembler は新しいMessageOutputAssemblerを作成します
func NewMessageOutputAssembler() *MessageOutputAssembler {
	return &MessageOutputAssembler{}
}

// AssembleMessageOutput はメッセージと関連データからMessageOutputを構築します
func (a *MessageOutputAssembler) AssembleMessageOutput(
	message *entity.Message,
	user *entity.User,
	userMentions []*entity.MessageUserMention,
	groupMentions []*entity.MessageGroupMention,
	links []*entity.MessageLink,
	reactions []*entity.MessageReaction,
	attachments []*entity.Attachment,
	groups map[string]*entity.UserGroup,
	userMap map[string]*entity.User,
) MessageOutput {
	userInfo := a.buildUserInfo(user)

	return MessageOutput{
		ID:          message.ID,
		ChannelID:   message.ChannelID,
		UserID:      message.UserID,
		User:        userInfo,
		ParentID:    message.ParentID,
		Body:        message.Body,
		Mentions:    a.buildUserMentions(userMentions),
		Groups:      a.buildGroupMentions(groupMentions, groups),
		Links:       a.buildLinks(links),
		Reactions:   a.buildReactions(reactions, userMap),
		Attachments: a.buildAttachments(attachments),
		CreatedAt:   message.CreatedAt,
		EditedAt:    message.EditedAt,
		DeletedAt:   message.DeletedAt,
	}
}

// buildUserInfo はユーザー情報を構築します
func (a *MessageOutputAssembler) buildUserInfo(user *entity.User) UserInfo {
	if user == nil {
		return UserInfo{
			ID:          "",
			DisplayName: "Unknown User",
			AvatarURL:   nil,
		}
	}

	return UserInfo{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
	}
}

// buildUserMentions はユーザーメンションを構築します
func (a *MessageOutputAssembler) buildUserMentions(userMentions []*entity.MessageUserMention) []UserMention {
	mentions := make([]UserMention, 0, len(userMentions))
	for _, mention := range userMentions {
		mentions = append(mentions, UserMention{
			UserID:      mention.UserID,
			DisplayName: "", // 必要に応じてユーザー情報を取得
		})
	}
	return mentions
}

// buildGroupMentions はグループメンションを構築します
func (a *MessageOutputAssembler) buildGroupMentions(groupMentions []*entity.MessageGroupMention, groups map[string]*entity.UserGroup) []GroupMention {
	groupMentionsOutput := make([]GroupMention, 0, len(groupMentions))
	for _, mention := range groupMentions {
		groupName := ""
		if group, exists := groups[mention.GroupID]; exists {
			groupName = group.Name
		}
		groupMentionsOutput = append(groupMentionsOutput, GroupMention{
			GroupID: mention.GroupID,
			Name:    groupName,
		})
	}
	return groupMentionsOutput
}

// buildLinks はリンク情報を構築します
func (a *MessageOutputAssembler) buildLinks(links []*entity.MessageLink) []LinkInfo {
	linksOutput := make([]LinkInfo, 0, len(links))
	for _, link := range links {
		linksOutput = append(linksOutput, LinkInfo{
			ID:          link.ID,
			URL:         link.URL,
			Title:       link.Title,
			Description: link.Description,
			ImageURL:    link.ImageURL,
			SiteName:    link.SiteName,
			CardType:    link.CardType,
		})
	}
	return linksOutput
}

// buildReactions はリアクション情報を構築します
func (a *MessageOutputAssembler) buildReactions(reactions []*entity.MessageReaction, userMap map[string]*entity.User) []ReactionInfo {
	reactionsOutput := make([]ReactionInfo, 0, len(reactions))
	for _, reaction := range reactions {
		reactionUser, exists := userMap[reaction.UserID]
		reactionUserInfo := UserInfo{
			ID:          reaction.UserID,
			DisplayName: "Unknown User",
			AvatarURL:   nil,
		}
		if exists && reactionUser != nil {
			reactionUserInfo = UserInfo{
				ID:          reactionUser.ID,
				DisplayName: reactionUser.DisplayName,
				AvatarURL:   reactionUser.AvatarURL,
			}
		}
		reactionsOutput = append(reactionsOutput, ReactionInfo{
			User:      reactionUserInfo,
			Emoji:     reaction.Emoji,
			CreatedAt: reaction.CreatedAt,
		})
	}
	return reactionsOutput
}

// buildAttachments は添付ファイル情報を構築します
func (a *MessageOutputAssembler) buildAttachments(attachments []*entity.Attachment) []AttachmentInfo {
	attachmentsOutput := make([]AttachmentInfo, 0, len(attachments))
	for _, attachment := range attachments {
		attachmentsOutput = append(attachmentsOutput, AttachmentInfo{
			ID:        attachment.ID,
			FileName:  attachment.FileName,
			MimeType:  attachment.MimeType,
			SizeBytes: attachment.SizeBytes,
		})
	}
	return attachmentsOutput
}

type MessageUseCase interface {
	ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error)
	CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error)
	UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error)
	DeleteMessage(ctx context.Context, input DeleteMessageInput) error
	GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error)
	GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error)
	ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error)
}

type messageInteractor struct {
	messageRepo           domainrepository.MessageRepository
	channelRepo           domainrepository.ChannelRepository
	channelMemberRepo     domainrepository.ChannelMemberRepository
	workspaceRepo         domainrepository.WorkspaceRepository
	userRepo              domainrepository.UserRepository
	userGroupRepo         domainrepository.UserGroupRepository
	userMentionRepo       domainrepository.MessageUserMentionRepository
	groupMentionRepo      domainrepository.MessageGroupMentionRepository
	linkRepo              domainrepository.MessageLinkRepository
	threadRepo            domainrepository.ThreadRepository
	attachmentRepo        domainrepository.AttachmentRepository
	ogpService            service.OGPService
	notificationSvc       service.NotificationService
	mentionService        service.MentionService
	linkProcessingService service.LinkProcessingService
	transactionManager    transaction.Manager
	assembler             *MessageOutputAssembler
}

func NewMessageUseCase(
	messageRepo domainrepository.MessageRepository,
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	userGroupRepo domainrepository.UserGroupRepository,
	userMentionRepo domainrepository.MessageUserMentionRepository,
	groupMentionRepo domainrepository.MessageGroupMentionRepository,
	linkRepo domainrepository.MessageLinkRepository,
	threadRepo domainrepository.ThreadRepository,
	attachmentRepo domainrepository.AttachmentRepository,
	ogpService service.OGPService,
	notificationSvc service.NotificationService,
	mentionService service.MentionService,
	linkProcessingService service.LinkProcessingService,
	transactionManager transaction.Manager,
) MessageUseCase {
	return &messageInteractor{
		messageRepo:           messageRepo,
		channelRepo:           channelRepo,
		channelMemberRepo:     channelMemberRepo,
		workspaceRepo:         workspaceRepo,
		userRepo:              userRepo,
		userGroupRepo:         userGroupRepo,
		userMentionRepo:       userMentionRepo,
		groupMentionRepo:      groupMentionRepo,
		linkRepo:              linkRepo,
		threadRepo:            threadRepo,
		attachmentRepo:        attachmentRepo,
		ogpService:            ogpService,
		notificationSvc:       notificationSvc,
		mentionService:        mentionService,
		linkProcessingService: linkProcessingService,
		transactionManager:    transactionManager,
		assembler:             NewMessageOutputAssembler(),
	}
}

func (i *messageInteractor) ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error) {
	channel, err := i.ensureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// リミット正規化
	limit := input.Limit
	if limit <= 0 {
		limit = defaultMessageLimit
	} else if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	messages, err := i.messageRepo.FindByChannelID(ctx, channel.ID, limit+1, input.Since, input.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	// メッセージリストの準備（リミット処理とID抽出）
	messageIDs, hasMore := i.prepareMessageList(messages, limit)

	// 関連データを一括取得
	relatedData, err := i.fetchRelatedData(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	// ユーザー情報を取得
	userMap, err := i.fetchUserMap(ctx, messages, relatedData.Reactions)
	if err != nil {
		return nil, err
	}

	// グループ情報を取得
	groups, err := i.fetchGroups(ctx, relatedData.GroupMentions)
	if err != nil {
		return nil, err
	}

	// メッセージ出力を構築
	outputs := i.buildMessageOutputs(messages, relatedData, userMap, groups)

	return &ListMessagesOutput{Messages: outputs, HasMore: hasMore}, nil
}

// prepareMessageList はメッセージリストを準備し、リミット処理とID抽出を行います
func (i *messageInteractor) prepareMessageList(messages []*entity.Message, limit int) ([]string, bool) {
	// リミット正規化
	if limit <= 0 {
		limit = defaultMessageLimit
	} else if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	// メッセージ切り詰め
	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	// メッセージID抽出
	messageIDs := make([]string, len(messages))
	for idx, msg := range messages {
		messageIDs[idx] = msg.ID
	}

	return messageIDs, hasMore
}

// fetchRelatedData はメッセージに関連するデータを一括取得します
func (i *messageInteractor) fetchRelatedData(ctx context.Context, messageIDs []string) (*RelatedData, error) {
	userMentions, err := i.userMentionRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user mentions: %w", err)
	}

	groupMentions, err := i.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group mentions: %w", err)
	}

	links, err := i.linkRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch links: %w", err)
	}

	reactions, err := i.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reactions: %w", err)
	}

	attachments, err := i.attachmentRepo.FindByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attachments: %w", err)
	}

	return &RelatedData{
		UserMentions:  userMentions,
		GroupMentions: groupMentions,
		Links:         links,
		Reactions:     reactions,
		Attachments:   attachments,
	}, nil
}

// fetchUserMap はユーザー情報を取得してマップに格納します
func (i *messageInteractor) fetchUserMap(ctx context.Context, messages []*entity.Message, reactions map[string][]*entity.MessageReaction) (map[string]*entity.User, error) {
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)

	// メッセージ作成者のユーザーIDを収集
	for _, msg := range messages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
		}
	}

	// リアクションユーザーIDも収集
	for _, reactionList := range reactions {
		for _, reaction := range reactionList {
			if !userIDSet[reaction.UserID] {
				userIDs = append(userIDs, reaction.UserID)
				userIDSet[reaction.UserID] = true
			}
		}
	}

	// ユーザー情報を一括取得
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	return userMap, nil
}

// fetchGroups はグループ情報を取得します
func (i *messageInteractor) fetchGroups(ctx context.Context, groupMentions []*entity.MessageGroupMention) (map[string]*entity.UserGroup, error) {
	groupIDs := make([]string, 0)
	groupIDSet := make(map[string]bool)

	for _, mention := range groupMentions {
		if !groupIDSet[mention.GroupID] {
			groupIDs = append(groupIDs, mention.GroupID)
			groupIDSet[mention.GroupID] = true
		}
	}

	if len(groupIDs) == 0 {
		return make(map[string]*entity.UserGroup), nil
	}

	// 一括でグループ情報を取得
	groups, err := i.userGroupRepo.FindByIDs(ctx, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch groups: %w", err)
	}

	// グループ情報をマップに格納
	groupMap := make(map[string]*entity.UserGroup)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	return groupMap, nil
}

// buildMessageOutputs はメッセージ出力を構築します
func (i *messageInteractor) buildMessageOutputs(messages []*entity.Message, relatedData *RelatedData, userMap map[string]*entity.User, groups map[string]*entity.UserGroup) []MessageOutput {
	// メンション、リンク、リアクションをメッセージIDでグループ化
	userMentionsByMessage := i.groupUserMentionsByMessage(relatedData.UserMentions)
	groupMentionsByMessage := i.groupGroupMentionsByMessage(relatedData.GroupMentions)
	linksByMessage := i.groupLinksByMessage(relatedData.Links)

	outputs := make([]MessageOutput, 0, len(messages))
	for _, msg := range messages {
		user := userMap[msg.UserID]
		outputs = append(outputs, i.assembler.AssembleMessageOutput(
			msg,
			user,
			userMentionsByMessage[msg.ID],
			groupMentionsByMessage[msg.ID],
			linksByMessage[msg.ID],
			relatedData.Reactions[msg.ID],
			relatedData.Attachments[msg.ID],
			groups,
			userMap,
		))
	}

	return outputs
}

// groupUserMentionsByMessage はユーザーメンションをメッセージIDでグループ化します
func (i *messageInteractor) groupUserMentionsByMessage(userMentions []*entity.MessageUserMention) map[string][]*entity.MessageUserMention {
	userMentionsByMessage := make(map[string][]*entity.MessageUserMention)
	for _, mention := range userMentions {
		userMentionsByMessage[mention.MessageID] = append(userMentionsByMessage[mention.MessageID], mention)
	}
	return userMentionsByMessage
}

// groupGroupMentionsByMessage はグループメンションをメッセージIDでグループ化します
func (i *messageInteractor) groupGroupMentionsByMessage(groupMentions []*entity.MessageGroupMention) map[string][]*entity.MessageGroupMention {
	groupMentionsByMessage := make(map[string][]*entity.MessageGroupMention)
	for _, mention := range groupMentions {
		groupMentionsByMessage[mention.MessageID] = append(groupMentionsByMessage[mention.MessageID], mention)
	}
	return groupMentionsByMessage
}

// groupLinksByMessage はリンクをメッセージIDでグループ化します
func (i *messageInteractor) groupLinksByMessage(links []*entity.MessageLink) map[string][]*entity.MessageLink {
	linksByMessage := make(map[string][]*entity.MessageLink)
	for _, link := range links {
		linksByMessage[link.MessageID] = append(linksByMessage[link.MessageID], link)
	}
	return linksByMessage
}

func (i *messageInteractor) CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error) {
	channel, err := i.ensureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	if input.ParentID != nil {
		parent, err := i.messageRepo.FindByID(ctx, *input.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch parent message: %w", err)
		}
		if parent == nil || parent.ChannelID != channel.ID {
			return nil, ErrParentMessageNotFound
		}
	}

	var result *MessageOutput
	err = i.transactionManager.Do(ctx, func(txCtx context.Context) error {
		message := &entity.Message{
			ChannelID: channel.ID,
			UserID:    input.UserID,
			ParentID:  input.ParentID,
			Body:      input.Body,
			CreatedAt: time.Now(),
		}

		if err := i.messageRepo.Create(txCtx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}

		// 添付ファイルをメッセージに紐付け
		if len(input.AttachmentIDs) > 0 {
			if err := i.attachmentRepo.AttachToMessage(txCtx, input.AttachmentIDs, message.ID); err != nil {
				return fmt.Errorf("failed to attach files: %w", err)
			}
		}

		// スレッド返信の場合、メタデータを更新
		if input.ParentID != nil {
			if err := i.threadRepo.IncrementReplyCount(txCtx, *input.ParentID, input.UserID); err != nil {
				return fmt.Errorf("failed to update thread metadata: %w", err)
			}
		}

		// メンションとリンクを抽出・保存
		if err := i.extractAndSaveMentionsAndLinks(txCtx, message.ID, input.Body, channel.WorkspaceID); err != nil {
			return fmt.Errorf("failed to extract mentions and links: %w", err)
		}

		// ユーザー情報を取得
		user, err := i.userRepo.FindByID(txCtx, input.UserID)
		if err != nil {
			return fmt.Errorf("failed to fetch user: %w", err)
		}

		// メンションとリンクの情報を取得してレスポンスに含める
		userMentions, err := i.userMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user mentions: %w", err)
		}
		groupMentions, err := i.groupMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch group mentions: %w", err)
		}
		links, err := i.linkRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch links: %w", err)
		}
		attachmentList, err := i.attachmentRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch attachments: %w", err)
		}

		// グループ情報を取得
		groupIDs := make([]string, 0)
		groupIDSet := make(map[string]bool)
		for _, mention := range groupMentions {
			if !groupIDSet[mention.GroupID] {
				groupIDs = append(groupIDs, mention.GroupID)
				groupIDSet[mention.GroupID] = true
			}
		}

		groups := make(map[string]*entity.UserGroup)
		if len(groupIDs) > 0 {
			groupList, err := i.userGroupRepo.FindByIDs(txCtx, groupIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch groups: %w", err)
			}
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}

		// リアクションは新規作成メッセージには存在しないため空配列
		reactions := []*entity.MessageReaction{}

		// ユーザーマップを作成
		userMap := map[string]*entity.User{user.ID: user}

		output := i.assembler.AssembleMessageOutput(message, user, userMentions, groupMentions, links, reactions, attachmentList, groups, userMap)
		result = &output

		return nil
	})

	if err != nil {
		return nil, err
	}

	// WebSocket通知を送信（nilチェックを追加）
	if i.notificationSvc != nil {
		// outputをmap[string]interface{}に変換
		messageMap, err := convertStructToMap(*result)
		if err == nil {
			i.notificationSvc.NotifyNewMessage(channel.WorkspaceID, channel.ID, messageMap)
		} else {
			logger.Get().Warn("Failed to convert message to map", zap.Error(err))
		}
	}

	return result, nil
}

func (i *messageInteractor) ensureChannelAccess(ctx context.Context, channelID, userID string) (*entity.Channel, error) {
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to load channel: %w", err)
	}
	if ch == nil {
		return nil, ErrChannelNotFound
	}

	if ch.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, ch.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify channel membership: %w", err)
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
		return ch, nil
	}

	member, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	return ch, nil
}

// メンションとリンクの抽出・保存
func (i *messageInteractor) extractAndSaveMentionsAndLinks(ctx context.Context, messageID, body, workspaceID string) error {
	// ユーザーメンションの抽出
	userMentions, err := i.mentionService.ExtractUserMentions(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to extract user mentions: %w", err)
	}
	for _, mention := range userMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := i.userMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create user mention: %w", err)
		}
	}

	// グループメンションの抽出
	groupMentions, err := i.mentionService.ExtractGroupMentions(ctx, body, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to extract group mentions: %w", err)
	}
	for _, mention := range groupMentions {
		mention.MessageID = messageID
		mention.CreatedAt = time.Now()
		if err := i.groupMentionRepo.Create(ctx, mention); err != nil {
			return fmt.Errorf("failed to create group mention: %w", err)
		}
	}

	// リンクの抽出とOGP取得
	links, err := i.linkProcessingService.ProcessLinks(ctx, body)
	if err != nil {
		return fmt.Errorf("failed to process links: %w", err)
	}

	for _, link := range links {
		// 既存のリンクをチェック
		existingLink, err := i.linkRepo.FindByURL(ctx, link.URL)
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
		if err := i.linkRepo.Create(ctx, link); err != nil {
			return fmt.Errorf("failed to create link: %w", err)
		}
	}

	return nil
}

func (i *messageInteractor) GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error) {
	// 親メッセージを取得
	parentMessage, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parent message: %w", err)
	}
	if parentMessage == nil {
		return nil, ErrParentMessageNotFound
	}

	// チャンネルアクセス権限を確認
	_, err = i.ensureChannelAccess(ctx, parentMessage.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// スレッド返信を取得
	replies, err := i.messageRepo.FindThreadReplies(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	// 親メッセージとリプライのメッセージIDを収集
	allMessages := append([]*entity.Message{parentMessage}, replies...)
	messageIDs := make([]string, len(allMessages))
	for idx, msg := range allMessages {
		messageIDs[idx] = msg.ID
	}

	// メンション、リンク、リアクション、添付ファイルを一括取得
	userMentions, _ := i.userMentionRepo.FindByMessageIDs(ctx, messageIDs)
	groupMentions, _ := i.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	links, _ := i.linkRepo.FindByMessageIDs(ctx, messageIDs)
	reactions, _ := i.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	attachments, _ := i.attachmentRepo.FindByMessageIDs(ctx, messageIDs)

	// ユーザーIDを収集
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	for _, msg := range allMessages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
		}
	}
	for _, reactionList := range reactions {
		for _, reaction := range reactionList {
			if !userIDSet[reaction.UserID] {
				userIDs = append(userIDs, reaction.UserID)
				userIDSet[reaction.UserID] = true
			}
		}
	}

	// ユーザー情報を一括取得
	users, _ := i.userRepo.FindByIDs(ctx, userIDs)
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メンション、リンク、リアクションをメッセージIDでグループ化
	userMentionsByMessage := make(map[string][]*entity.MessageUserMention)
	for _, mention := range userMentions {
		userMentionsByMessage[mention.MessageID] = append(userMentionsByMessage[mention.MessageID], mention)
	}

	groupMentionsByMessage := make(map[string][]*entity.MessageGroupMention)
	for _, mention := range groupMentions {
		groupMentionsByMessage[mention.MessageID] = append(groupMentionsByMessage[mention.MessageID], mention)
	}

	linksByMessage := make(map[string][]*entity.MessageLink)
	for _, link := range links {
		linksByMessage[link.MessageID] = append(linksByMessage[link.MessageID], link)
	}

	// グループ情報を取得
	groupIDs := make([]string, 0)
	groupIDSet := make(map[string]bool)
	for _, mention := range groupMentions {
		if !groupIDSet[mention.GroupID] {
			groupIDs = append(groupIDs, mention.GroupID)
			groupIDSet[mention.GroupID] = true
		}
	}

	groups := make(map[string]*entity.UserGroup)
	if len(groupIDs) > 0 {
		groupList, err := i.userGroupRepo.FindByIDs(ctx, groupIDs)
		if err == nil {
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}
	}

	// 親メッセージ出力を作成
	parentOutput := i.assembler.AssembleMessageOutput(
		parentMessage,
		userMap[parentMessage.UserID],
		userMentionsByMessage[parentMessage.ID],
		groupMentionsByMessage[parentMessage.ID],
		linksByMessage[parentMessage.ID],
		reactions[parentMessage.ID],
		attachments[parentMessage.ID],
		groups,
		userMap,
	)

	// リプライ出力を作成
	replyOutputs := make([]MessageOutput, 0, len(replies))
	for _, reply := range replies {
		replyOutputs = append(replyOutputs, i.assembler.AssembleMessageOutput(
			reply,
			userMap[reply.UserID],
			userMentionsByMessage[reply.ID],
			groupMentionsByMessage[reply.ID],
			linksByMessage[reply.ID],
			reactions[reply.ID],
			attachments[reply.ID],
			groups,
			userMap,
		))
	}

	return &GetThreadRepliesOutput{
		ParentMessage: parentOutput,
		Replies:       replyOutputs,
		HasMore:       false,
	}, nil
}

func (i *messageInteractor) GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error) {
	// メッセージの存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, ErrParentMessageNotFound
	}

	// チャンネルアクセス権限を確認
	_, err = i.ensureChannelAccess(ctx, message.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// スレッドメタデータを取得
	metadata, err := i.threadRepo.FindMetadataByMessageID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread metadata: %w", err)
	}

	// メタデータが存在しない場合は空のメタデータを返す
	if metadata == nil {
		return &ThreadMetadataOutput{
			MessageID:          input.MessageID,
			ReplyCount:         0,
			LastReplyAt:        nil,
			LastReplyUser:      nil,
			ParticipantUserIDs: []string{},
		}, nil
	}

	// 最新返信者の情報を取得
	var lastReplyUser *UserInfo
	if metadata.LastReplyUserID != nil {
		user, err := i.userRepo.FindByID(ctx, *metadata.LastReplyUserID)
		if err == nil && user != nil {
			lastReplyUser = &UserInfo{
				ID:          user.ID,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			}
		}
	}

	return &ThreadMetadataOutput{
		MessageID:          metadata.MessageID,
		ReplyCount:         metadata.ReplyCount,
		LastReplyAt:        metadata.LastReplyAt,
		LastReplyUser:      lastReplyUser,
		ParticipantUserIDs: metadata.ParticipantUserIDs,
	}, nil
}

// convertStructToMap は構造体をmap[string]interface{}に変換します
// より効率的なreflectパッケージを使用した実装
func convertStructToMap(data interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", v.Kind())
	}

	t := v.Type()
	result := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 非公開フィールドはスキップ
		if !field.CanInterface() {
			continue
		}

		// JSONタグからフィールド名を取得、なければ構造体フィールド名を使用
		jsonTag := fieldType.Tag.Get("json")
		fieldName := fieldType.Name
		if jsonTag != "" && jsonTag != "-" {
			// カンマ以降を除去（omitempty等のオプションを除去）
			if commaIndex := strings.Index(jsonTag, ","); commaIndex > 0 {
				fieldName = jsonTag[:commaIndex]
			} else {
				fieldName = jsonTag
			}
		}

		// フィールドがnilポインタでない場合のみ追加
		if field.IsValid() && !field.IsZero() {
			result[fieldName] = field.Interface()
		}
	}

	return result, nil
}

func (i *messageInteractor) ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error) {
	// 通常のメッセージ一覧を取得
	listOutput, err := i.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}

	// メッセージIDを収集
	messageIDs := make([]string, len(listOutput.Messages))
	for idx, msg := range listOutput.Messages {
		messageIDs[idx] = msg.ID
	}

	// スレッドメタデータを一括取得
	metadataMap, err := i.threadRepo.FindMetadataByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread metadata: %w", err)
	}

	// 最新返信者のユーザーIDを収集
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	for _, metadata := range metadataMap {
		if metadata.LastReplyUserID != nil && !userIDSet[*metadata.LastReplyUserID] {
			userIDs = append(userIDs, *metadata.LastReplyUserID)
			userIDSet[*metadata.LastReplyUserID] = true
		}
	}

	// ユーザー情報を一括取得
	users, _ := i.userRepo.FindByIDs(ctx, userIDs)
	userMap := make(map[string]*entity.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メッセージとスレッドメタデータを結合
	outputs := make([]MessageWithThreadOutput, 0, len(listOutput.Messages))
	for _, msg := range listOutput.Messages {
		output := MessageWithThreadOutput{
			MessageOutput: msg,
		}

		if metadata, exists := metadataMap[msg.ID]; exists {
			var lastReplyUser *UserInfo
			if metadata.LastReplyUserID != nil {
				user := userMap[*metadata.LastReplyUserID]
				if user != nil {
					lastReplyUser = &UserInfo{
						ID:          user.ID,
						DisplayName: user.DisplayName,
						AvatarURL:   user.AvatarURL,
					}
				}
			}

			output.ThreadMetadata = &ThreadMetadataOutput{
				MessageID:          metadata.MessageID,
				ReplyCount:         metadata.ReplyCount,
				LastReplyAt:        metadata.LastReplyAt,
				LastReplyUser:      lastReplyUser,
				ParticipantUserIDs: metadata.ParticipantUserIDs,
			}
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

func (i *messageInteractor) UpdateMessage(ctx context.Context, input UpdateMessageInput) (*MessageOutput, error) {
	// メッセージ存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("メッセージの取得に失敗しました: %w", err)
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}

	// チャンネルアクセス確認
	channel, err := i.ensureChannelAccess(ctx, message.ChannelID, input.EditorID)
	if err != nil {
		return nil, err
	}

	// 削除済みメッセージの編集禁止
	if message.DeletedAt != nil {
		return nil, ErrCannotEditDeleted
	}

	// 権限確認: 投稿者本人または管理者
	canEdit, err := i.canModifyMessage(ctx, channel.WorkspaceID, message.UserID, input.EditorID)
	if err != nil {
		return nil, fmt.Errorf("権限確認に失敗しました: %w", err)
	}
	if !canEdit {
		return nil, ErrUnauthorized
	}

	var result *MessageOutput
	err = i.transactionManager.Do(ctx, func(txCtx context.Context) error {
		// メッセージ本文を更新
		message.Body = input.Body
		now := time.Now()
		message.EditedAt = &now

		// データベース更新
		if err := i.messageRepo.Update(txCtx, message); err != nil {
			return fmt.Errorf("メッセージの更新に失敗しました: %w", err)
		}

		// 既存のメンション・リンクを削除
		if err := i.userMentionRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete user mentions: %w", err)
		}
		if err := i.groupMentionRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete group mentions: %w", err)
		}
		if err := i.linkRepo.DeleteByMessageID(txCtx, message.ID); err != nil {
			return fmt.Errorf("failed to delete links: %w", err)
		}

		// 新しいメンション・リンクを抽出・保存
		if err := i.extractAndSaveMentionsAndLinks(txCtx, message.ID, input.Body, channel.WorkspaceID); err != nil {
			return fmt.Errorf("failed to extract and save mentions/links: %w", err)
		}

		// 更新後のデータを取得してMessageOutputを構築
		userMentions, err := i.userMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch user mentions: %w", err)
		}

		groupMentions, err := i.groupMentionRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch group mentions: %w", err)
		}

		links, err := i.linkRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch links: %w", err)
		}

		reactions, err := i.messageRepo.FindReactions(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch reactions: %w", err)
		}

		attachmentList, err := i.attachmentRepo.FindByMessageID(txCtx, message.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch attachments: %w", err)
		}

		// ユーザー情報を取得
		user, err := i.userRepo.FindByID(txCtx, message.UserID)
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
			groupList, err := i.userGroupRepo.FindByIDs(txCtx, groupIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch groups: %w", err)
			}
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}

		userMap := map[string]*entity.User{user.ID: user}
		output := i.assembler.AssembleMessageOutput(message, user, userMentions, groupMentions, links, reactions, attachmentList, groups, userMap)
		result = &output

		return nil
	})

	if err != nil {
		return nil, err
	}

	// WebSocket通知を送信
	if i.notificationSvc != nil {
		messageMap, err := convertStructToMap(*result)
		if err == nil {
			i.notificationSvc.NotifyUpdatedMessage(channel.WorkspaceID, channel.ID, messageMap)
		} else {
			logger.Get().Warn("Failed to convert message to map", zap.Error(err))
		}
	}

	return result, nil
}

func (i *messageInteractor) DeleteMessage(ctx context.Context, input DeleteMessageInput) error {
	// メッセージ存在確認
	message, err := i.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return fmt.Errorf("メッセージの取得に失敗しました: %w", err)
	}
	if message == nil {
		return ErrMessageNotFound
	}

	// チャンネルアクセス確認
	channel, err := i.ensureChannelAccess(ctx, message.ChannelID, input.ExecutorID)
	if err != nil {
		return err
	}

	// 既に削除済みの場合はエラー
	if message.DeletedAt != nil {
		return ErrMessageAlreadyDeleted
	}

	// 権限確認: 投稿者本人または管理者
	canDelete, err := i.canModifyMessage(ctx, channel.WorkspaceID, message.UserID, input.ExecutorID)
	if err != nil {
		return fmt.Errorf("権限確認に失敗しました: %w", err)
	}
	if !canDelete {
		return ErrUnauthorized
	}

	// 削除対象メッセージIDのリストを作成
	deleteIDs := []string{message.ID}

	// スレッド親メッセージの場合、子メッセージも削除
	if message.ParentID == nil {
		replies, err := i.messageRepo.FindThreadReplies(ctx, message.ID)
		if err != nil {
			return fmt.Errorf("返信の取得に失敗しました: %w", err)
		}
		for _, reply := range replies {
			deleteIDs = append(deleteIDs, reply.ID)
		}

		// スレッドメタデータも削除
		if err := i.threadRepo.DeleteMetadata(ctx, message.ID); err != nil {
			logger.Get().Warn("Failed to delete thread metadata", zap.Error(err))
		}
	}

	// ソフトデリート実行
	if err := i.messageRepo.SoftDeleteByIDs(ctx, deleteIDs, input.ExecutorID); err != nil {
		return fmt.Errorf("メッセージの削除に失敗しました: %w", err)
	}

	// WebSocket通知を送信
	if i.notificationSvc != nil {
		deleteData := map[string]interface{}{
			"messageId":  message.ID,
			"channelId":  message.ChannelID,
			"deletedIds": deleteIDs,
		}
		i.notificationSvc.NotifyDeletedMessage(channel.WorkspaceID, channel.ID, deleteData)
	}

	return nil
}

// canModifyMessage はユーザーがメッセージを編集・削除できるかどうかを確認します
func (i *messageInteractor) canModifyMessage(ctx context.Context, workspaceID, messageOwnerID, executorID string) (bool, error) {
	// 投稿者本人の場合は許可
	if messageOwnerID == executorID {
		return true, nil
	}

	// 管理者権限チェック
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, executorID)
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
