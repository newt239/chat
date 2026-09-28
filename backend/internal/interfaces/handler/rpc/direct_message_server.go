package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	dmuc "github.com/newt239/chat/internal/usecase/dm"
)

type DirectMessageServer struct {
	UC *dmuc.Interactor
}

func (s *DirectMessageServer) ListDirectMessages(ctx context.Context, req *chatv1.ListDirectMessagesRequest) (*chatv1.ListDirectMessagesResponse, error) {
	out, err := s.UC.ListDMs(ctx, dmuc.ListDMsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.ListDirectMessagesResponse{DirectMessages: presenter.ConvertAll(out, presenter.DirectMessage)}, nil
}

func (s *DirectMessageServer) CreateDirectMessage(ctx context.Context, req *chatv1.CreateDirectMessageRequest) (*chatv1.CreateDirectMessageResponse, error) {
	out, err := s.UC.CreateDM(ctx, dmuc.CreateDMInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), TargetUserID: req.UserId})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateDirectMessageResponse{DirectMessage: presenter.DirectMessage(out)}, nil
}

func (s *DirectMessageServer) CreateGroupDirectMessage(ctx context.Context, req *chatv1.CreateGroupDirectMessageRequest) (*chatv1.CreateGroupDirectMessageResponse, error) {
	out, err := s.UC.CreateGroupDM(ctx, dmuc.CreateGroupDMInput{
		WorkspaceID: req.WorkspaceId,
		CreatorID:   userIDFrom(ctx),
		MemberIDs:   req.UserIds,
		Name:        req.GetName(),
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.CreateGroupDirectMessageResponse{DirectMessage: presenter.DirectMessage(out)}, nil
}
