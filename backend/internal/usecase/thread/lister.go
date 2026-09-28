package thread

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/message"
)

type ThreadLister struct {
	threadRepo           domainrepository.ThreadRepository
	messageOutputBuilder *message.MessageOutputBuilder
}

func NewThreadLister(
	threadRepo domainrepository.ThreadRepository,
	messageOutputBuilder *message.MessageOutputBuilder,
) *ThreadLister {
	return &ThreadLister{
		threadRepo:           threadRepo,
		messageOutputBuilder: messageOutputBuilder,
	}
}

func (l *ThreadLister) ListParticipatingThreads(ctx context.Context, input ListParticipatingThreadsInput) (*ListParticipatingThreadsOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	limit = min(limit, 100)

	result, err := l.threadRepo.FindParticipatingThreads(ctx, domainrepository.FindParticipatingThreadsInput{
		WorkspaceID:          input.WorkspaceID,
		UserID:               input.UserID,
		CursorLastActivityAt: input.CursorLastActivityAt,
		CursorThreadID:       input.CursorThreadID,
		Limit:                limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find participating threads: %w", err)
	}

	// 親と最新の返信をまとめて組み立てる
	messages := []*entity.Message{}
	for _, item := range result.Items {
		messages = append(messages, item.FirstMessage)
		messages = append(messages, item.LatestReplies...)
	}
	outputs, err := l.messageOutputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build message outputs: %w", err)
	}

	items := make([]ParticipatingThreadOutput, 0, len(result.Items))
	for _, item := range result.Items {
		first := outputs[0]
		replyCount := len(item.LatestReplies)
		items = append(items, ParticipatingThreadOutput{
			ThreadID:       item.ThreadID,
			ChannelID:      item.ChannelID,
			FirstMessage:   &first,
			LatestReplies:  outputs[1 : 1+replyCount],
			ReplyCount:     item.ReplyCount,
			LastActivityAt: item.LastActivityAt,
			UnreadCount:    item.UnreadCount,
		})
		outputs = outputs[1+replyCount:]
	}

	var nextCursor *ThreadCursorOutput
	if result.NextCursor != nil {
		nextCursor = &ThreadCursorOutput{
			LastActivityAt: result.NextCursor.LastActivityAt,
			ThreadID:       result.NextCursor.ThreadID,
		}
	}

	return &ListParticipatingThreadsOutput{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}
