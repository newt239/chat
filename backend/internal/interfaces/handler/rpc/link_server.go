package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	linkuc "github.com/newt239/chat/internal/usecase/link"
)

type LinkServer struct {
	UC linkuc.LinkUseCase
}

func (s *LinkServer) FetchOgp(ctx context.Context, req *chatv1.FetchOgpRequest) (*chatv1.FetchOgpResponse, error) {
	out, err := s.UC.FetchOGP(ctx, linkuc.FetchOGPInput{URL: req.Url})
	if err != nil {
		return nil, err
	}
	return &chatv1.FetchOgpResponse{Ogp: presenter.OGPData(out.OGPData)}, nil
}
