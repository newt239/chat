package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	useruc "github.com/newt239/chat/internal/usecase/user"
)

type UserServer struct {
	UC useruc.UseCase
}

func (s *UserServer) GetMe(ctx context.Context, _ *chatv1.GetMeRequest) (*chatv1.GetMeResponse, error) {
	out, err := s.UC.GetMe(ctx, userIDFrom(ctx))
	if err != nil {
		return nil, err
	}
	return &chatv1.GetMeResponse{User: presenter.Me(out)}, nil
}

func (s *UserServer) UpdateMe(ctx context.Context, req *chatv1.UpdateMeRequest) (*chatv1.UpdateMeResponse, error) {
	out, err := s.UC.UpdateMe(ctx, useruc.UpdateMeInput{
		UserID:      userIDFrom(ctx),
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		AvatarURL:   req.AvatarUrl,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateMeResponse{User: presenter.Me(out)}, nil
}

func (s *UserServer) UpdatePassword(ctx context.Context, req *chatv1.UpdatePasswordRequest) (*chatv1.UpdatePasswordResponse, error) {
	err := s.UC.UpdatePassword(ctx, useruc.UpdatePasswordInput{
		UserID:          userIDFrom(ctx),
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdatePasswordResponse{}, nil
}

func (s *UserServer) DeleteMe(ctx context.Context, _ *chatv1.DeleteMeRequest) (*chatv1.DeleteMeResponse, error) {
	if err := s.UC.DeleteMe(ctx, userIDFrom(ctx)); err != nil {
		return nil, err
	}
	return &chatv1.DeleteMeResponse{}, nil
}
