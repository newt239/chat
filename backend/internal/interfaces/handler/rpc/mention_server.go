package rpc

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type MentionServer struct {
	UC *messageuc.Interactor
}

func (s *MentionServer) ListMentions(ctx context.Context, req *chatv1.ListMentionsRequest) (*chatv1.ListMentionsResponse, error) {
	input := domainrepository.FindMentionsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), Limit: int(req.Limit)}
	if c := req.Cursor; c != nil {
		input.Cursor = &domainrepository.MessageCursor{CreatedAt: c.CreatedAt.AsTime(), MessageID: c.MessageId}
	}
	out, err := s.UC.ListMentions(ctx, input)
	if err != nil {
		return nil, err
	}
	return presenter.Mentions(out), nil
}
