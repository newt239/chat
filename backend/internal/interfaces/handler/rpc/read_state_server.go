package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
)

type ReadStateServer struct {
	UC readstateuc.ReadStateUseCase
}

func (s *ReadStateServer) UpdateReadState(ctx context.Context, req *chatv1.UpdateReadStateRequest) (*chatv1.UpdateReadStateResponse, error) {
	err := s.UC.UpdateReadState(ctx, readstateuc.UpdateReadStateInput{
		ChannelID:          req.ChannelId,
		UserID:             userIDFrom(ctx),
		LastReadAt:         req.LastReadAt.AsTime(),
		IncludeDescendants: req.IncludeDescendants,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateReadStateResponse{}, nil
}

func (s *ReadStateServer) GetUnreadCount(ctx context.Context, req *chatv1.GetUnreadCountRequest) (*chatv1.GetUnreadCountResponse, error) {
	out, err := s.UC.GetUnreadCount(ctx, readstateuc.GetUnreadCountInput{ChannelID: req.ChannelId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetUnreadCountResponse{Count: int32(out.Count)}, nil
}
