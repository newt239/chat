package message

import (
	"context"
	"fmt"
	"sort"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

// MessageLister はメッセージ一覧取得を担当するユースケースです
type MessageLister struct {
	messageRepo      domainrepository.MessageRepository
	systemMsgRepo    domainrepository.SystemMessageRepository
	userRepo         domainrepository.UserRepository
	threadRepo       domainrepository.ThreadRepository
	outputBuilder    *MessageOutputBuilder
	channelAccessSvc service.ChannelAccessService
}

// NewMessageLister は新しいMessageListerを作成します
func NewMessageLister(
	messageRepo domainrepository.MessageRepository,
	systemMsgRepo domainrepository.SystemMessageRepository,
	userRepo domainrepository.UserRepository,
	threadRepo domainrepository.ThreadRepository,
	outputBuilder *MessageOutputBuilder,
	channelAccessSvc service.ChannelAccessService,
) *MessageLister {
	return &MessageLister{
		messageRepo:      messageRepo,
		systemMsgRepo:    systemMsgRepo,
		userRepo:         userRepo,
		threadRepo:       threadRepo,
		outputBuilder:    outputBuilder,
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

	// システムメッセージ取得
	systemMessages, err := l.systemMsgRepo.FindByChannelID(ctx, channel.ID, limit+1, input.Since, input.Until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch system messages: %w", err)
	}

	// ユーザーメッセージの出力へ変換
	messages, hasMoreUser := l.prepareMessageList(messages, limit)
	userOutputs, err := l.outputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, err
	}

	// タイムラインへマージ
	timeline := make([]TimelineItem, 0, len(userOutputs)+len(systemMessages))
	for _, m := range userOutputs {
		timeline = append(timeline, TimelineItem{Type: "user", UserMessage: &m, CreatedAt: m.CreatedAt})
	}
	for _, sm := range systemMessages {
		timeline = append(timeline, TimelineItem{Type: "system", SystemMessage: &SystemMessageOutput{
			ID:        sm.ID,
			ChannelID: sm.ChannelID,
			Kind:      string(sm.Kind),
			Payload:   sm.Payload,
			ActorID:   sm.ActorID,
			CreatedAt: sm.CreatedAt,
		}, CreatedAt: sm.CreatedAt})
	}
	sort.Slice(timeline, func(i, j int) bool { return timeline[i].CreatedAt.After(timeline[j].CreatedAt) })
	hasMore := false
	if len(timeline) > limit {
		hasMore = true
		timeline = timeline[:limit]
	}

	return &ListMessagesOutput{Messages: timeline, HasMore: hasMore || hasMoreUser}, nil
}

// ListMessagesWithThread はスレッド情報付きのメッセージ一覧を取得します
func (l *MessageLister) ListMessagesWithThread(ctx context.Context, input ListMessagesInput) ([]MessageWithThreadOutput, error) {
	// 通常のメッセージ一覧を取得（統合タイムライン）
	listOutput, err := l.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}

	// ユーザーメッセージのみ抽出しID収集
	userMessages := make([]MessageOutput, 0)
	messageIDs := make([]string, 0)
	for _, item := range listOutput.Messages {
		if item.Type == "user" && item.UserMessage != nil {
			userMessages = append(userMessages, *item.UserMessage)
			messageIDs = append(messageIDs, item.UserMessage.ID)
		}
	}

	// スレッドメタデータを一括計算
	metadataMap, err := l.threadRepo.CalculateMetadataByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate thread metadata: %w", err)
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
	outputs := make([]MessageWithThreadOutput, 0, len(userMessages))
	for _, msg := range userMessages {
		output := MessageWithThreadOutput{MessageOutput: msg}

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

	outputs, err := l.outputBuilder.Build(ctx, input.UserID, append([]*entity.Message{parentMessage}, replies...))
	if err != nil {
		return nil, err
	}

	return &GetThreadRepliesOutput{
		ParentMessage: outputs[0],
		Replies:       outputs[1:],
		HasMore:       false,
	}, nil
}

// GetMessagePreview はメッセージリンクの引用カードを取得します
func (l *MessageLister) GetMessagePreview(ctx context.Context, input GetMessagePreviewInput) (*MessagePreviewOutput, error) {
	return l.outputBuilder.BuildPreview(ctx, input.UserID, input.MessageID)
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

	// スレッドメタデータを計算
	metadata, err := l.threadRepo.CalculateMetadataByMessageID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate thread metadata: %w", err)
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
