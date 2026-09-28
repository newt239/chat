package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	mentionuc "github.com/newt239/chat/internal/usecase/mention"
)

type MentionServer struct {
	Lister *mentionuc.Lister
}

func (s *MentionServer) ListMentions(ctx context.Context, req *chatv1.ListMentionsRequest) (*chatv1.ListMentionsResponse, error) {
	input := mentionuc.ListMentionsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), Limit: int(req.Limit)}
	if req.Cursor != nil {
		input.Cursor = &mentionuc.Cursor{CreatedAt: req.Cursor.CreatedAt.AsTime(), MessageID: req.Cursor.MessageId}
	}
	out, err := s.Lister.ListMentions(ctx, input)
	if err != nil {
		return nil, err
	}
	return presenter.Mentions(out), nil
}
