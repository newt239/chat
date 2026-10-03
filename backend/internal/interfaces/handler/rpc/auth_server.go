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

// handOverRefreshToken はブラウザには Cookie で渡し、ネイティブアプリにだけ本文で返すリフレッシュトークンを返します
func handOverRefreshToken(ctx context.Context, out *authuc.AuthOutput) *string {
	if isNativeClient(ctx) {
		return &out.RefreshToken
	}
	setRefreshTokenCookie(ctx, out.RefreshToken, out.ExpiresAt)
	return nil
}

func (s *AuthServer) GetAuthConfig(context.Context, *chatv1.GetAuthConfigRequest) (*chatv1.GetAuthConfigResponse, error) {
	return &chatv1.GetAuthConfigResponse{PasswordAuthEnabled: s.UC.PasswordAuthEnabled()}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *chatv1.LoginRequest) (*chatv1.LoginResponse, error) {
	out, err := s.UC.Login(ctx, authuc.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &chatv1.LoginResponse{AccessToken: out.AccessToken, RefreshToken: handOverRefreshToken(ctx, out), User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) LoginWithGoogle(ctx context.Context, req *chatv1.LoginWithGoogleRequest) (*chatv1.LoginWithGoogleResponse, error) {
	out, err := s.UC.LoginWithGoogle(ctx, authuc.LoginWithGoogleInput{IDToken: req.IdToken, WorkspaceID: req.WorkspaceId})
	if err != nil {
		return nil, err
	}
	return &chatv1.LoginWithGoogleResponse{AccessToken: out.AccessToken, RefreshToken: handOverRefreshToken(ctx, out), User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) LoginWithGoogleCode(ctx context.Context, req *chatv1.LoginWithGoogleCodeRequest) (*chatv1.LoginWithGoogleCodeResponse, error) {
	out, err := s.UC.LoginWithGoogleCode(ctx, authuc.LoginWithGoogleCodeInput{Code: req.Code, CodeVerifier: req.CodeVerifier, Nonce: req.Nonce, WorkspaceID: req.WorkspaceId})
	if err != nil {
		return nil, err
	}
	return &chatv1.LoginWithGoogleCodeResponse{AccessToken: out.AccessToken, RefreshToken: handOverRefreshToken(ctx, out), User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) SignUp(ctx context.Context, req *chatv1.SignUpRequest) (*chatv1.SignUpResponse, error) {
	out, err := s.UC.SignUp(ctx, authuc.SignUpInput{WorkspaceID: req.WorkspaceId, Email: req.Email, DisplayName: req.DisplayName, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &chatv1.SignUpResponse{AccessToken: out.AccessToken, RefreshToken: handOverRefreshToken(ctx, out), User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) SignUpWithInvitation(ctx context.Context, req *chatv1.SignUpWithInvitationRequest) (*chatv1.SignUpWithInvitationResponse, error) {
	out, err := s.UC.SignUpWithInvitation(ctx, authuc.SignUpWithInvitationInput{Token: req.Token, DisplayName: req.DisplayName, Password: req.Password})
	if err != nil {
		return nil, err
	}
	return &chatv1.SignUpWithInvitationResponse{AccessToken: out.AccessToken, RefreshToken: handOverRefreshToken(ctx, out), User: presenter.AuthUser(out.User)}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *chatv1.RefreshRequest) (*chatv1.RefreshResponse, error) {
	token := req.GetRefreshToken()
	if !isNativeClient(ctx) {
		token = refreshTokenFromCookie(ctx)
	}
	out, err := s.UC.RefreshToken(ctx, authuc.RefreshTokenInput{RefreshToken: token})
	if err != nil {
		return nil, err
	}
	return &chatv1.RefreshResponse{AccessToken: out.AccessToken, RefreshToken: handOverRefreshToken(ctx, out), User: presenter.AuthUser(out.User)}, nil
}

// Logout は公開 RPC で、Cookie のリフレッシュトークンか、あればアクセストークンのセッションを失効させます
func (s *AuthServer) Logout(ctx context.Context, _ *chatv1.LogoutRequest) (*chatv1.LogoutResponse, error) {
	clearRefreshTokenCookie(ctx)
	input := authuc.LogoutInput{RefreshToken: refreshTokenFromCookie(ctx), SessionID: claimsFrom(ctx).SessionID}
	if err := s.UC.Logout(ctx, input); err != nil {
		return nil, err
	}
	return &chatv1.LogoutResponse{}, nil
}
