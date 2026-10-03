package message

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// ListMessages はメッセージ一覧を取得します
func (i *Interactor) ListMessages(ctx context.Context, input ListMessagesInput) (*ListMessagesOutput, error) {
	channel, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}

	channelIDs := []string{channel.ID}
	if input.IncludeDescendants {
		descendants, err := i.channelAccessSvc.AccessibleDescendants(ctx, channel, input.UserID)
		if err != nil {
			return nil, err
		}
		for _, d := range descendants {
			channelIDs = append(channelIDs, d.ID)
		}
	}

	if input.Around != nil {
		older, hasMore, err := i.fetchTimeline(ctx, input.UserID, channelIDs, input.Limit, nil, input.Around, false)
		if err != nil {
			return nil, err
		}
		// 指定日時ちょうどの投稿も後ろ側に含める。created_at はマイクロ秒精度
		since := input.Around.Add(-time.Microsecond)
		newer, hasNewer, err := i.fetchTimeline(ctx, input.UserID, channelIDs, input.Limit, &since, nil, true)
		if err != nil {
			return nil, err
		}
		slices.Reverse(newer)
		return &ListMessagesOutput{Messages: append(newer, older...), HasMore: hasMore, HasNewer: hasNewer}, nil
	}

	// since だけの指定は続きの読み込みなので、since の直後から古い順に取る
	if input.Since != nil && input.Until == nil {
		newer, hasNewer, err := i.fetchTimeline(ctx, input.UserID, channelIDs, input.Limit, input.Since, nil, true)
		if err != nil {
			return nil, err
		}
		slices.Reverse(newer)
		return &ListMessagesOutput{Messages: newer, HasNewer: hasNewer}, nil
	}

	timeline, hasMore, err := i.fetchTimeline(ctx, input.UserID, channelIDs, input.Limit, input.Since, input.Until, false)
	if err != nil {
		return nil, err
	}
	return &ListMessagesOutput{Messages: timeline, HasMore: hasMore}, nil
}

// fetchTimeline はユーザー・システムメッセージを合わせて limit 件取り、取得した向きに続きがあるかを返します
func (i *Interactor) fetchTimeline(ctx context.Context, userID string, channelIDs []string, limit int, since, until *time.Time, ascending bool) ([]TimelineItem, bool, error) {
	messages, err := i.messageRepo.FindByChannelIDs(ctx, channelIDs, limit+1, since, until, ascending)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch messages: %w", err)
	}
	systemMessages, err := i.systemMsgRepo.FindByChannelIDs(ctx, channelIDs, limit+1, since, until, ascending)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch system messages: %w", err)
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	userOutputs, err := i.outputBuilder.Build(ctx, userID, messages)
	if err != nil {
		return nil, false, err
	}
	messageIDs := make([]string, len(messages))
	for idx, m := range messages {
		messageIDs[idx] = m.ID
	}
	threads, err := i.buildThreadMetadata(ctx, userID, messageIDs)
	if err != nil {
		return nil, false, err
	}

	timeline := make([]TimelineItem, 0, len(userOutputs)+len(systemMessages))
	for _, m := range userOutputs {
		if t := threads[m.ID]; t.ReplyCount > 0 {
			m.ThreadMetadata = t
		}
		timeline = append(timeline, TimelineItem{UserMessage: &m, CreatedAt: m.CreatedAt})
	}
	for _, sm := range systemMessages {
		timeline = append(timeline, TimelineItem{SystemMessage: sm, CreatedAt: sm.CreatedAt})
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

// buildThreadMetadata はメッセージごとのスレッドの返信数・最新の返信者・閲覧者のフォロー状態をまとめて求めます
func (i *Interactor) buildThreadMetadata(ctx context.Context, userID string, messageIDs []string) (map[string]*ThreadMetadataOutput, error) {
	metadataMap, err := i.threadRepo.CalculateMetadataByMessageIDs(ctx, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate thread metadata: %w", err)
	}
	followed, err := i.threadRepo.FindFollowedThreadIDs(ctx, userID, messageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find followed threads: %w", err)
	}

	var replierIDs []string
	for _, metadata := range metadataMap {
		if metadata.LastReplyUserID != nil {
			replierIDs = append(replierIDs, *metadata.LastReplyUserID)
		}
	}
	repliers, err := i.userRepo.FindByIDs(ctx, replierIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load last repliers: %w", err)
	}

	result := make(map[string]*ThreadMetadataOutput, len(metadataMap))
	for id, metadata := range metadataMap {
		out := &ThreadMetadataOutput{
			ReplyCount:  metadata.ReplyCount,
			LastReplyAt: metadata.LastReplyAt,
			IsFollowing: followed[id],
		}
		if metadata.LastReplyUserID != nil {
			info := UserInfoOf(*metadata.LastReplyUserID, repliers)
			out.LastReplyUser = &info
		}
		result[id] = out
	}
	return result, nil
}

// GetThreadReplies はスレッド返信を取得します
func (i *Interactor) GetThreadReplies(ctx context.Context, input GetThreadRepliesInput) (*GetThreadRepliesOutput, error) {
	parentMessage, _, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID)
	if err != nil {
		return nil, err
	}

	replies, hasMore, hasNewer, err := i.fetchThreadReplies(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch thread replies: %w", err)
	}

	outputs, err := i.outputBuilder.Build(ctx, input.UserID, append([]*entity.Message{parentMessage}, replies...))
	if err != nil {
		return nil, err
	}
	metadata, err := i.threadRepo.CalculateMetadataByMessageIDs(ctx, []string{input.MessageID})
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
func (i *Interactor) fetchThreadReplies(ctx context.Context, input GetThreadRepliesInput) (replies []*entity.Message, hasMore, hasNewer bool, err error) {
	page := func(since, until *time.Time, ascending bool) ([]*entity.Message, bool, error) {
		found, err := i.messageRepo.FindThreadReplies(ctx, input.MessageID, input.Limit+1, since, until, ascending)
		if err != nil || len(found) <= input.Limit {
			return found, false, err
		}
		return found[:input.Limit], true, nil
	}

	since, until := input.Since, input.Until
	if input.AroundReplyID != nil {
		target, err := i.messageRepo.FindByID(ctx, *input.AroundReplyID)
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
func (i *Interactor) GetMessagePreview(ctx context.Context, input MessageInput) (*MessagePreviewOutput, error) {
	return i.outputBuilder.BuildPreview(ctx, input.UserID, input.MessageID)
}

// GetThreadMetadata はスレッドの情報を閲覧者のフォロー状態付きで返します
func (i *Interactor) GetThreadMetadata(ctx context.Context, input MessageInput) (*ThreadMetadataOutput, error) {
	if _, _, err := i.channelAccessSvc.EnsureMessageAccess(ctx, input.MessageID, input.UserID); err != nil {
		return nil, err
	}

	metadata, err := i.buildThreadMetadata(ctx, input.UserID, []string{input.MessageID})
	if err != nil {
		return nil, err
	}
	return metadata[input.MessageID], nil
}
