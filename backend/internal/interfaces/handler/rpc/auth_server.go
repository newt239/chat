package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type AuthServer struct {
	UC authuc.AuthUseCase
}

func (s *AuthServer) Register(ctx context.Context, req *chatv1.RegisterRequest) (*chatv1.RegisterResponse, error) {
	out, err := s.UC.Register(ctx, authuc.RegisterInput{Email: req.Email, Password: req.Password, DisplayName: req.DisplayName})
	if err != nil {
		return nil, err
	}
	return &chatv1.RegisterResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken, User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *chatv1.LoginRequest) (*chatv1.LoginResponse, error) {
	out, err := s.UC.Login(ctx, authuc.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &chatv1.LoginResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken, User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *chatv1.RefreshRequest) (*chatv1.RefreshResponse, error) {
	out, err := s.UC.RefreshToken(ctx, authuc.RefreshTokenInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return nil, err
	}
	return &chatv1.RefreshResponse{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken, User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) Logout(ctx context.Context, _ *chatv1.LogoutRequest) (*chatv1.LogoutResponse, error) {
	if _, err := s.UC.Logout(ctx, authuc.LogoutInput{UserID: userIDFrom(ctx)}); err != nil {
		return nil, err
	}
	return &chatv1.LogoutResponse{}, nil
}
