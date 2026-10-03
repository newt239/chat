package mention

import (
	"cmp"
	"context"
	"fmt"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type ListMentionsOutput struct {
	Messages   []messageuc.MessageOutput
	NextCursor *domainrepository.MessageCursor
}

type Interactor struct {
	workspaceRepo        domainrepository.WorkspaceRepository
	messageRepo          domainrepository.MessageRepository
	messageOutputBuilder *messageuc.MessageOutputBuilder
}

func New(
	workspaceRepo domainrepository.WorkspaceRepository,
	messageRepo domainrepository.MessageRepository,
	messageOutputBuilder *messageuc.MessageOutputBuilder,
) *Interactor {
	return &Interactor{workspaceRepo: workspaceRepo, messageRepo: messageRepo, messageOutputBuilder: messageOutputBuilder}
}

// ListMentions は自分宛てのメンションを含むメッセージを新しい順に返します
func (i *Interactor) ListMentions(ctx context.Context, input domainrepository.FindMentionsInput) (*ListMentionsOutput, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, input.WorkspaceID, input.UserID); err != nil {
		return nil, err
	}
	limit := min(cmp.Or(input.Limit, defaultLimit), maxLimit)
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
	outputs, err := i.messageOutputBuilder.Build(ctx, input.UserID, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to build message outputs: %w", err)
	}
	return &ListMentionsOutput{Messages: outputs, NextCursor: nextCursor}, nil
}
