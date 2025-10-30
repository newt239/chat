package message

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
    domainservice "github.com/newt239/chat/internal/domain/service"
)

// MessageLister はメッセージ一覧取得を担当するユースケースです
type MessageLister struct {
	messageRepo       domainrepository.MessageRepository
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	userRepo          domainrepository.UserRepository
	userGroupRepo     domainrepository.UserGroupRepository
	userMentionRepo   domainrepository.MessageUserMentionRepository
	groupMentionRepo  domainrepository.MessageGroupMentionRepository
	linkRepo          domainrepository.MessageLinkRepository
	threadRepo        domainrepository.ThreadRepository
	attachmentRepo    domainrepository.AttachmentRepository
	assembler         *MessageOutputAssembler
	outputBuilder     *MessageOutputBuilder
    channelAccessSvc  domainservice.ChannelAccessService
}

// NewMessageLister は新しいMessageListerを作成します
func NewMessageLister(
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
    channelAccessSvc domainservice.ChannelAccessService,
) *MessageLister {
	assembler := NewMessageOutputAssembler()
	return &MessageLister{
		messageRepo:       messageRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
		userGroupRepo:     userGroupRepo,
		userMentionRepo:   userMentionRepo,
		groupMentionRepo:  groupMentionRepo,
		linkRepo:          linkRepo,
		threadRepo:        threadRepo,
		attachmentRepo:    attachmentRepo,
		assembler:         assembler,
		outputBuilder: NewMessageOutputBuilder(
			messageRepo,
			userRepo,
			userGroupRepo,
			userMentionRepo,
			groupMentionRepo,
			linkRepo,
			attachmentRepo,
			assembler,
		),
        channelAccessSvc: channelAccessSvc,
	}
}

// ListMessages はメッセージ一覧を取得します
func (l *MessageLister) ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error) {
    channel, err := l.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
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

	messages, err := l.messageRepo.FindByChannelID(ctx, channel.ID, limit+1, input.Since, input.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	messages, hasMore := l.prepareMessageList(messages, limit)

	outputs, err := l.outputBuilder.Build(ctx, messages)
	if err != nil {
		return nil, err
	}

	return &ListMessagesOutput{Messages: outputs, HasMore: hasMore}, nil
}

// ListMessagesWithThread はスレッド情報付きのメッセージ一覧を取得します
func (l *MessageLister) ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error) {
	// 通常のメッセージ一覧を取得
	listOutput, err := l.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}

	// メッセージIDを収集
	messageIDs := make([]string, len(listOutput.Messages))
	for idx, msg := range listOutput.Messages {
		messageIDs[idx] = msg.ID
	}

	// スレッドメタデータを一括取得
	metadataMap, err := l.threadRepo.FindMetadataByMessageIDs(ctx, messageIDs)
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
	users, _ := l.userRepo.FindByIDs(ctx, userIDs)
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

// GetThreadReplies はスレッド返信を取得します
func (l *MessageLister) GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error) {
	// 親メッセージを取得
	parentMessage, err := l.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parent message: %w", err)
	}
	if parentMessage == nil {
		return nil, ErrParentMessageNotFound
	}

	// チャンネルアクセス権限を確認
    _, err = l.channelAccessSvc.EnsureChannelAccess(ctx, parentMessage.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// スレッド返信を取得
	replies, err := l.messageRepo.FindThreadReplies(ctx, input.MessageID)
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
	userMentions, _ := l.userMentionRepo.FindByMessageIDs(ctx, messageIDs)
	groupMentions, _ := l.groupMentionRepo.FindByMessageIDs(ctx, messageIDs)
	links, _ := l.linkRepo.FindByMessageIDs(ctx, messageIDs)
	reactions, _ := l.messageRepo.FindReactionsByMessageIDs(ctx, messageIDs)
	attachments, _ := l.attachmentRepo.FindByMessageIDs(ctx, messageIDs)

	// ユーザーIDを収集
	userIDs := make([]string, 0)
	userIDSet := make(map[string]bool)
	for _, msg := range allMessages {
		if !userIDSet[msg.UserID] {
			userIDs = append(userIDs, msg.UserID)
			userIDSet[msg.UserID] = true
		}
		if msg.DeletedBy != nil && !userIDSet[*msg.DeletedBy] {
			userIDs = append(userIDs, *msg.DeletedBy)
			userIDSet[*msg.DeletedBy] = true
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
	users, _ := l.userRepo.FindByIDs(ctx, userIDs)
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
		groupList, err := l.userGroupRepo.FindByIDs(ctx, groupIDs)
		if err == nil {
			for _, group := range groupList {
				groups[group.ID] = group
			}
		}
	}

	// 親メッセージ出力を作成
	parentOutput := l.assembler.AssembleMessageOutput(
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
		replyOutputs = append(replyOutputs, l.assembler.AssembleMessageOutput(
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

// GetThreadMetadata はスレッドメタデータを取得します
func (l *MessageLister) GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error) {
	// メッセージの存在確認
	message, err := l.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, ErrParentMessageNotFound
	}

	// チャンネルアクセス権限を確認
    _, err = l.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	// スレッドメタデータを取得
	metadata, err := l.threadRepo.FindMetadataByMessageID(ctx, input.MessageID)
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
		user, err := l.userRepo.FindByID(ctx, *metadata.LastReplyUserID)
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

// ensureChannelAccess はチャンネルアクセス権限を確認します
// ensureChannelAccess は ChannelAccessService に委譲済み

// prepareMessageList はメッセージリストを準備し、リミット処理を行います
func (l *MessageLister) prepareMessageList(messages []*entity.Message, limit int) ([]*entity.Message, bool) {
	if limit <= 0 {
		limit = defaultMessageLimit
	} else if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	hasMore := false
	if len(messages) > limit {
		hasMore = true
		messages = messages[:limit]
	}

	return messages, hasMore
}
