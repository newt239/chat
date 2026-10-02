package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	polluc "github.com/newt239/chat/internal/usecase/poll"
)

type PollServer struct {
	UC *polluc.Interactor
}

func (s *PollServer) Vote(ctx context.Context, req *chatv1.VoteRequest) (*chatv1.VoteResponse, error) {
	out, err := s.UC.Vote(ctx, polluc.VoteInput{PollID: req.PollId, UserID: userIDFrom(ctx), OptionIDs: req.OptionIds})
	if err != nil {
		return nil, err
	}
	return &chatv1.VoteResponse{Message: presenter.Message(*out)}, nil
}

func (s *PollServer) ClosePoll(ctx context.Context, req *chatv1.ClosePollRequest) (*chatv1.ClosePollResponse, error) {
	out, err := s.UC.Close(ctx, polluc.CloseInput{PollID: req.PollId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ClosePollResponse{Message: presenter.Message(*out)}, nil
}
