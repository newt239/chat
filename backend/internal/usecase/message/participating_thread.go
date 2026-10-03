package message

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

const defaultListLimit = 20

type ParticipatingThreadOutput struct {
	ThreadID       string
	ChannelID      string
	FirstMessage   MessageOutput
	LatestReplies  []MessageOutput
	ReplyCount     int
	LastActivityAt time.Time
	UnreadCount    int
	IsFollowing    bool
}

type ListParticipatingThreadsOutput struct {
	Items      []ParticipatingThreadOutput
	NextCursor *domainrepository.ThreadCursor
}

// ListParticipatingThreads はフォロー中・返信した・返信でメンションされたスレッドを最終アクティビティの新しい順に返します
func (i *Interactor) ListParticipatingThreads(ctx context.Context, input domainrepository.FindParticipatingThreadsInput) (*ListParticipatingThreadsOutput, error) {
	if _, err := service.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
	}
	input.Limit = min(cmp.Or(input.Limit, defaultListLimit), maxMessageLimit)
	result, err := i.threadRepo.FindParticipatingThreads(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to find participating threads: %w", err)
	}

	// 親と最新の返信をまとめて組み立てる
	var messages []*entity.Message
	threadIDs := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		messages = append(messages, item.FirstMessage)
		messages = append(messages, item.LatestReplies...)
		threadIDs = append(threadIDs, item.ThreadID)
	}
	followed, err := i.threadRepo.FindFollowedThreadIDs(ctx, input.UserID, threadIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find followed threads: %w", err)
	}
	outputs, err := i.outputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build message outputs: %w", err)
	}

	items := make([]ParticipatingThreadOutput, 0, len(result.Items))
	for _, item := range result.Items {
		replyCount := len(item.LatestReplies)
		items = append(items, ParticipatingThreadOutput{
			ThreadID:       item.ThreadID,
			ChannelID:      item.ChannelID,
			FirstMessage:   outputs[0],
			LatestReplies:  outputs[1 : 1+replyCount],
			ReplyCount:     item.ReplyCount,
			LastActivityAt: item.LastActivityAt,
			UnreadCount:    item.UnreadCount,
			IsFollowing:    followed[item.ThreadID],
		})
		outputs = outputs[1+replyCount:]
	}
	return &ListParticipatingThreadsOutput{Items: items, NextCursor: result.NextCursor}, nil
}

type ListMentionsOutput struct {
	Messages   []MessageOutput
	NextCursor *domainrepository.MessageCursor
}

// ListMentions は自分宛てのメンションを含むメッセージを新しい順に返します
func (i *Interactor) ListMentions(ctx context.Context, input domainrepository.FindMentionsInput) (*ListMentionsOutput, error) {
	if _, err := service.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
	}
	limit := min(cmp.Or(input.Limit, defaultListLimit), maxMessageLimit)
	input.Limit = limit + 1
	messages, err := i.messageRepo.FindMentions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to find mentions: %w", err)
	}

	var nextCursor *domainrepository.MessageCursor
	if len(messages) > limit {
		messages = messages[:limit]
		last := messages[limit-1]
		nextCursor = &domainrepository.MessageCursor{CreatedAt: last.CreatedAt, MessageID: last.ID}
	}
	outputs, err := i.outputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build message outputs: %w", err)
	}
	return &ListMentionsOutput{Messages: outputs, NextCursor: nextCursor}, nil
}

func (i *Interactor) MarkThreadRead(ctx context.Context, threadID, userID string) error {
	if _, _, err := i.channelAccessSvc.EnsureMessageAccess(ctx, threadID, userID); err != nil {
		return err
	}
	if err := i.threadRepo.UpsertReadState(ctx, userID, threadID, time.Now()); err != nil {
		return fmt.Errorf("failed to mark thread as read: %w", err)
	}
	return nil
}

// SetFollowing はスレッドのフォローを付け外しします。既にその状態でも成功します
func (i *Interactor) SetFollowing(ctx context.Context, threadID, userID string, following bool) error {
	if _, _, err := i.channelAccessSvc.EnsureMessageAccess(ctx, threadID, userID); err != nil {
		return err
	}
	if err := i.threadRepo.SetFollowing(ctx, userID, threadID, following); err != nil {
		return fmt.Errorf("failed to update thread follow: %w", err)
	}
	return nil
}
