package mention

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

var ErrUnauthorized = errors.New("このワークスペースのメンションを取得する権限がありません")

type Cursor struct {
	CreatedAt time.Time
	MessageID string
}

type ListMentionsInput struct {
	WorkspaceID string
	UserID      string
	Cursor      *Cursor
	Limit       int
}

type ListMentionsOutput struct {
	Messages   []messageuc.MessageOutput
	NextCursor *Cursor
}

type Lister struct {
	workspaceRepo        domainrepository.WorkspaceRepository
	messageRepo          domainrepository.MessageRepository
	messageOutputBuilder *messageuc.MessageOutputBuilder
}

func NewLister(
	workspaceRepo domainrepository.WorkspaceRepository,
	messageRepo domainrepository.MessageRepository,
	messageOutputBuilder *messageuc.MessageOutputBuilder,
) *Lister {
	return &Lister{workspaceRepo: workspaceRepo, messageRepo: messageRepo, messageOutputBuilder: messageOutputBuilder}
}

// ListMentions は自分宛てのメンションを含むメッセージを新しい順に返します
func (l *Lister) ListMentions(ctx context.Context, input ListMentionsInput) (*ListMentionsOutput, error) {
	member, err := l.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	limit := input.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	limit = min(limit, maxLimit)

	repoInput := domainrepository.FindMentionsInput{WorkspaceID: input.WorkspaceID, UserID: input.UserID, Limit: limit + 1}
	if input.Cursor != nil {
		repoInput.Cursor = &domainrepository.MessageCursor{CreatedAt: input.Cursor.CreatedAt, MessageID: input.Cursor.MessageID}
	}
	messages, err := l.messageRepo.FindMentions(ctx, repoInput)
	if err != nil {
		return nil, fmt.Errorf("failed to find mentions: %w", err)
	}

	var nextCursor *Cursor
	if len(messages) > limit {
		messages = messages[:limit]
		last := messages[limit-1]
		nextCursor = &Cursor{CreatedAt: last.CreatedAt, MessageID: last.ID}
	}

	outputs, err := l.messageOutputBuilder.Build(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build message outputs: %w", err)
	}
	return &ListMentionsOutput{Messages: outputs, NextCursor: nextCursor}, nil
}
