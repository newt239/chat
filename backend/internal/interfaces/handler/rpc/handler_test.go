package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	domerr "github.com/newt239/chat/internal/domain/errors"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

const validToken = "valid-token"

type fakeJWTService struct{}

func (fakeJWTService) GenerateToken(string, time.Duration) (string, error) { return validToken, nil }

func (fakeJWTService) VerifyToken(token string) (*authuc.TokenClaims, error) {
	if token != validToken {
		return nil, errors.New("invalid")
	}
	return &authuc.TokenClaims{UserID: "user-1"}, nil
}

type stubAuthServer struct {
	chatv1connect.UnimplementedAuthServiceHandler
	logoutErr error
}

func (s stubAuthServer) Login(context.Context, *chatv1.LoginRequest) (*chatv1.LoginResponse, error) {
	return &chatv1.LoginResponse{AccessToken: validToken}, nil
}

func (s stubAuthServer) Logout(ctx context.Context, _ *chatv1.LogoutRequest) (*chatv1.LogoutResponse, error) {
	if userIDFrom(ctx) != "user-1" {
		return nil, errors.New("ユーザー ID がコンテキストにありません")
	}
	return &chatv1.LogoutResponse{}, s.logoutErr
}

func newTestClient(t *testing.T, server stubAuthServer) chatv1connect.AuthServiceClient {
	t.Helper()
	ts := httptest.NewServer(NewHandler(fakeJWTService{}, Register(chatv1connect.NewAuthServiceHandler, chatv1connect.AuthServiceHandler(server))))
	t.Cleanup(ts.Close)
	return chatv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)
}

func withToken(token string) context.Context {
	ctx, callInfo := connect.NewClientContext(context.Background())
	callInfo.RequestHeader().Set("Authorization", "Bearer "+token)
	return ctx
}

func TestPublicProcedureDoesNotRequireToken(t *testing.T) {
	client := newTestClient(t, stubAuthServer{})
	if _, err := client.Login(context.Background(), &chatv1.LoginRequest{Email: "alice@example.com", Password: "password123"}); err != nil {
		t.Fatalf("公開 RPC がトークンなしで失敗しました: %v", err)
	}
}

func TestProtectedProcedureRequiresValidToken(t *testing.T) {
	client := newTestClient(t, stubAuthServer{})
	for name, ctx := range map[string]context.Context{
		"トークンなし":  context.Background(),
		"不正なトークン": withToken("invalid"),
	} {
		if _, err := client.Logout(ctx, &chatv1.LogoutRequest{}); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s: Unauthenticated を期待しましたが %v でした", name, err)
		}
	}
	if _, err := client.Logout(withToken(validToken), &chatv1.LogoutRequest{}); err != nil {
		t.Errorf("有効なトークンで失敗しました: %v", err)
	}
}

func TestInvalidRequestIsRejected(t *testing.T) {
	client := newTestClient(t, stubAuthServer{})
	_, err := client.Login(context.Background(), &chatv1.LoginRequest{Email: "not-an-email", Password: "x"})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("InvalidArgument を期待しましたが %v でした", err)
	}
}

func TestUseCaseErrorIsConverted(t *testing.T) {
	tests := []struct {
		err  error
		want connect.Code
	}{
		{fmt.Errorf("取得に失敗: %w", domerr.ErrMessageNotFound), connect.CodeNotFound},
		{domerr.ErrUnauthorized, connect.CodePermissionDenied},
		{errors.New("想定外"), connect.CodeInternal},
	}
	for _, tt := range tests {
		client := newTestClient(t, stubAuthServer{logoutErr: tt.err})
		_, err := client.Logout(withToken(validToken), &chatv1.LogoutRequest{})
		if connect.CodeOf(err) != tt.want {
			t.Errorf("%v: %v を期待しましたが %v でした", tt.err, tt.want, err)
		}
	}
}
