package message

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

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

	channelIDs := []string{channel.ID}
	if input.IncludeDescendants {
		descendants, err := l.channelAccessSvc.AccessibleDescendants(ctx, channel, input.UserID)
		if err != nil {
			return nil, err
		}
		for _, d := range descendants {
			channelIDs = append(channelIDs, d.ID)
		}
	}

	if input.Around != nil {
		older, hasMore, err := l.fetchTimeline(ctx, input.UserID, channelIDs, limit, nil, input.Around, false)
		if err != nil {
			return nil, err
		}
		// 指定日時ちょうどの投稿も後ろ側に含める。created_at はマイクロ秒精度
		since := input.Around.Add(-time.Microsecond)
		newer, hasNewer, err := l.fetchTimeline(ctx, input.UserID, channelIDs, limit, &since, nil, true)
		if err != nil {
			return nil, err
		}
		slices.Reverse(newer)
		return &ListMessagesOutput{Messages: append(newer, older...), HasMore: hasMore, HasNewer: hasNewer}, nil
	}

	// since だけの指定は続きの読み込みなので、since の直後から古い順に取る
	if input.Since != nil && input.Until == nil {
		newer, hasNewer, err := l.fetchTimeline(ctx, input.UserID, channelIDs, limit, input.Since, nil, true)
		if err != nil {
			return nil, err
		}
		slices.Reverse(newer)
		return &ListMessagesOutput{Messages: newer, HasNewer: hasNewer}, nil
	}

	timeline, hasMore, err := l.fetchTimeline(ctx, input.UserID, channelIDs, limit, input.Since, input.Until, false)
	if err != nil {
		return nil, err
	}
	return &ListMessagesOutput{Messages: timeline, HasMore: hasMore}, nil
}

// fetchTimeline はユーザー・システムメッセージを合わせて limit 件取り、取得した向きに続きがあるかを返します
func (l *MessageLister) fetchTimeline(ctx context.Context, userID string, channelIDs []string, limit int, since, until *time.Time, ascending bool) ([]TimelineItem, bool, error) {
	messages, err := l.messageRepo.FindByChannelIDs(ctx, channelIDs, limit+1, since, until, ascending)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch messages: %w", err)
	}
	systemMessages, err := l.systemMsgRepo.FindByChannelIDs(ctx, channelIDs, limit+1, since, until, ascending)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch system messages: %w", err)
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	userOutputs, err := l.outputBuilder.Build(ctx, userID, messages)
	if err != nil {
		return nil, false, err
	}

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
	sort.SliceStable(timeline, func(i, j int) bool {
		if ascending {
			return timeline[i].CreatedAt.Before(timeline[j].CreatedAt)
		}
		return timeline[i].CreatedAt.After(timeline[j].CreatedAt)
	})
	if len(timeline) > limit {
		hasMore = true
		timeline = timeline[:limit]
	}
	return timeline, hasMore, nil
}

// ListMessagesWithThread はユーザーのメッセージだけを、閲覧者から見たスレッドの情報付きで返します
func (l *MessageLister) ListMessagesWithThread(ctx context.Context, input ListMessagesInput) (*ListMessagesWithThreadOutput, error) {
	list, err := l.ListMessages(ctx, input)
	if err != nil {
		return nil, err
	}

	messages := make([]MessageOutput, 0, len(list.Messages))
	messageIDs := make([]string, 0, len(list.Messages))
	for _, item := range list.Messages {
		if item.UserMessage != nil {
			messages = append(messages, *item.UserMessage)
			messageIDs = append(messageIDs, item.UserMessage.ID)
		}
	}
	metadata, err := l.buildThreadMetadata(ctx, input.UserID, messageIDs)
	if err != nil {
		return nil, err
	}

	outputs := make([]MessageWithThreadOutput, 0, len(messages))
	for _, msg := range messages {
		outputs = append(outputs, MessageWithThreadOutput{MessageOutput: msg, ThreadMetadata: metadata[msg.ID]})
	}
	return &ListMessagesWithThreadOutput{Messages: outputs, HasMore: list.HasMore}, nil
}

// buildThreadMetadata はメッセージごとのスレッドの返信数・最新の返信者・閲覧者のフォロー状態をまとめて求めます
func (l *MessageLister) buildThreadMetadata(ctx context.Context, userID string, messageIDs []string) (map[string]*ThreadMetadataOutput, error) {
	metadataMap, err := l.threadRepo.CalculateMetadataByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate thread metadata: %w", err)
	}
	followed, err := l.threadRepo.FindFollowedThreadIDs(ctx, userID, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find followed threads: %w", err)
	}

	var replierIDs []string
	for _, metadata := range metadataMap {
		if metadata.LastReplyUserID != nil && !slices.Contains(replierIDs, *metadata.LastReplyUserID) {
			replierIDs = append(replierIDs, *metadata.LastReplyUserID)
		}
	}
	repliers, err := l.userRepo.FindByIDs(ctx, replierIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load last repliers: %w", err)
	}
	replierMap := make(map[string]*entity.User, len(repliers))
	for _, u := range repliers {
		replierMap[u.ID] = u
	}

	result := make(map[string]*ThreadMetadataOutput, len(metadataMap))
	for id, metadata := range metadataMap {
		out := &ThreadMetadataOutput{
			MessageID:   id,
			ReplyCount:  metadata.ReplyCount,
			LastReplyAt: metadata.LastReplyAt,
			IsFollowing: followed[id],
		}
		if metadata.LastReplyUserID != nil {
			out.LastReplyUser = new(toUserInfo(*metadata.LastReplyUserID, replierMap))
		}
		result[id] = out
	}
	return result, nil
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

	replies, hasMore, hasNewer, err := l.fetchThreadReplies(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	outputs, err := l.outputBuilder.Build(ctx, input.UserID, append([]*entity.Message{parentMessage}, replies...))
	if err != nil {
		return nil, err
	}
	metadata, err := l.threadRepo.CalculateMetadataByMessageIDs(ctx, []string{input.MessageID})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate thread metadata: %w", err)
	}

	return &GetThreadRepliesOutput{
		ParentMessage: outputs[0],
		Replies:       outputs[1:],
		HasMore:       hasMore,
		HasNewer:      hasNewer,
		ReplyCount:    metadata[input.MessageID].ReplyCount,
	}, nil
}

// fetchThreadReplies は返信を古い順に limit 件ずつ取り、前後に続きがあるかを返します。範囲の指定がなければ最新の返信を返します
func (l *MessageLister) fetchThreadReplies(ctx context.Context, input GetThreadRepliesInput) (replies []*entity.Message, hasMore, hasNewer bool, err error) {
	limit := input.Limit
	if limit <= 0 {
		limit = defaultMessageLimit
	} else if limit > maxMessageLimit {
		limit = maxMessageLimit
	}
	page := func(since, until *time.Time, ascending bool) ([]*entity.Message, bool, error) {
		found, err := l.messageRepo.FindThreadReplies(ctx, input.MessageID, limit+1, since, until, ascending)
		if err != nil || len(found) <= limit {
			return found, false, err
		}
		return found[:limit], true, nil
	}

	since, until := input.Since, input.Until
	if input.AroundReplyID != nil {
		target, err := l.messageRepo.FindByID(ctx, *input.AroundReplyID)
		if err != nil {
			return nil, false, false, err
		}
		if target != nil && target.ParentID != nil && *target.ParentID == input.MessageID {
			// 指定した返信そのものも後ろ側に含める。created_at はマイクロ秒精度
			since, until = new(target.CreatedAt.Add(-time.Microsecond)), new(target.CreatedAt)
		}
	}

	var older, newer []*entity.Message
	if until != nil || since == nil {
		if older, hasMore, err = page(nil, until, false); err != nil {
			return nil, false, false, err
		}
		slices.Reverse(older)
	}
	if since != nil {
		if newer, hasNewer, err = page(since, nil, true); err != nil {
			return nil, false, false, err
		}
	}
	return append(older, newer...), hasMore, hasNewer, nil
}

// GetMessagePreview はメッセージリンクの引用カードを取得します
func (l *MessageLister) GetMessagePreview(ctx context.Context, input GetMessagePreviewInput) (*MessagePreviewOutput, error) {
	return l.outputBuilder.BuildPreview(ctx, input.UserID, input.MessageID)
}

// GetThreadMetadata はスレッドの情報を閲覧者のフォロー状態付きで返します
func (l *MessageLister) GetThreadMetadata(ctx context.Context, input GetThreadMetadataInput) (*ThreadMetadataOutput, error) {
	message, err := l.messageRepo.FindByID(ctx, input.MessageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch message: %w", err)
	}
	if message == nil {
		return nil, ErrParentMessageNotFound
	}
	if _, err := l.channelAccessSvc.EnsureChannelAccess(ctx, message.ChannelID, input.UserID); err != nil {
		return nil, err
	}

	metadata, err := l.buildThreadMetadata(ctx, input.UserID, []string{input.MessageID})
	if err != nil {
		return nil, err
	}
	return metadata[input.MessageID], nil
}
