package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
)

type ReadStateServer struct {
	UC *readstateuc.Interactor
}

func (s *ReadStateServer) UpdateReadState(ctx context.Context, req *chatv1.UpdateReadStateRequest) (*chatv1.UpdateReadStateResponse, error) {
	return &chatv1.UpdateReadStateResponse{}, s.UC.UpdateReadState(ctx, readstateuc.UpdateReadStateInput{
		ChannelID:          req.ChannelId,
		UserID:             userIDFrom(ctx),
		LastReadAt:         req.LastReadAt.AsTime(),
		IncludeDescendants: req.IncludeDescendants,
	})
}
