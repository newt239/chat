package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usernoteuc "github.com/newt239/chat/internal/usecase/usernote"
)

type UserServer struct {
	UC     useruc.UseCase
	NoteUC usernoteuc.UseCase
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

func (s *UserServer) GetUserNote(ctx context.Context, req *chatv1.GetUserNoteRequest) (*chatv1.GetUserNoteResponse, error) {
	out, err := s.NoteUC.Get(ctx, usernoteuc.GetInput{OwnerID: userIDFrom(ctx), TargetID: req.TargetUserId})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetUserNoteResponse{Note: presenter.UserNote(out)}, nil
}

func (s *UserServer) UpdateUserNote(ctx context.Context, req *chatv1.UpdateUserNoteRequest) (*chatv1.UpdateUserNoteResponse, error) {
	out, err := s.NoteUC.Update(ctx, usernoteuc.UpdateInput{
		OwnerID:  userIDFrom(ctx),
		TargetID: req.TargetUserId,
		Nickname: req.Nickname,
		Memo:     req.Memo,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.UpdateUserNoteResponse{Note: presenter.UserNote(out)}, nil
}
