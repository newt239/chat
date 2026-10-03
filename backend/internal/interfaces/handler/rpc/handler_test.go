package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	domerr "github.com/newt239/chat/internal/domain/errors"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

const (
	validToken    = "valid-token"
	allowedOrigin = "https://chat.localhost"
)

type fakeJWTService struct{}

func (fakeJWTService) GenerateToken(authuc.TokenClaims, time.Duration) (string, error) {
	return validToken, nil
}

func (fakeJWTService) VerifyToken(token string) (*authuc.TokenClaims, error) {
	if token != validToken {
		return nil, errors.New("invalid")
	}
	return &authuc.TokenClaims{UserID: "user-1", SessionID: "session-1"}, nil
}

// stubAuthUseCase は渡された入力を記録し、固定のトークンを返します
type stubAuthUseCase struct {
	authUseCase
	refreshed []string
	loggedOut []authuc.LogoutInput
}

func (u *stubAuthUseCase) Login(context.Context, authuc.LoginInput) (*authuc.AuthOutput, error) {
	return &authuc.AuthOutput{AccessToken: validToken, RefreshToken: "rt", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (u *stubAuthUseCase) RefreshToken(_ context.Context, input authuc.RefreshTokenInput) (*authuc.AuthOutput, error) {
	u.refreshed = append(u.refreshed, input.RefreshToken)
	return &authuc.AuthOutput{AccessToken: validToken, RefreshToken: "rotated", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (u *stubAuthUseCase) Logout(_ context.Context, input authuc.LogoutInput) error {
	u.loggedOut = append(u.loggedOut, input)
	return nil
}

type stubRealtimeServer struct {
	chatv1connect.UnimplementedRealtimeServiceHandler
	err error
}

func (s stubRealtimeServer) IssueWebSocketTicket(ctx context.Context, _ *chatv1.IssueWebSocketTicketRequest) (*chatv1.IssueWebSocketTicketResponse, error) {
	if userIDFrom(ctx) != "user-1" {
		return nil, errors.New("ユーザー ID がコンテキストにありません")
	}
	return &chatv1.IssueWebSocketTicketResponse{Ticket: "ticket"}, s.err
}

func newTestServer(t *testing.T, uc *stubAuthUseCase, realtime stubRealtimeServer) *httptest.Server {
	t.Helper()
	opts := HandlerOptions(fakeJWTService{}, []string{allowedOrigin})
	mux := http.NewServeMux()
	mux.Handle(chatv1connect.NewAuthServiceHandler(&AuthServer{UC: uc}, opts...))
	mux.Handle(chatv1connect.NewRealtimeServiceHandler(realtime, opts...))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func headers(pairs ...string) context.Context {
	ctx, _ := withHeaders(pairs...)
	return ctx
}

func withHeaders(pairs ...string) (context.Context, connect.CallInfo) {
	ctx, callInfo := connect.NewClientContext(context.Background())
	for i := 0; i < len(pairs); i += 2 {
		callInfo.RequestHeader().Set(pairs[i], pairs[i+1])
	}
	return ctx, callInfo
}

const wsRequest = `{"workspaceId":"00000000-0000-0000-0000-000000000001"}`

func issueTicket(client chatv1connect.RealtimeServiceClient, ctx context.Context) error {
	_, err := client.IssueWebSocketTicket(ctx, &chatv1.IssueWebSocketTicketRequest{WorkspaceId: "00000000-0000-0000-0000-000000000001"})
	return err
}

func TestProtectedProcedureRequiresValidToken(t *testing.T) {
	ts := newTestServer(t, &stubAuthUseCase{}, stubRealtimeServer{})
	client := chatv1connect.NewRealtimeServiceClient(http.DefaultClient, ts.URL)
	for name, ctx := range map[string]context.Context{
		"トークンなし":  context.Background(),
		"不正なトークン": headers("Authorization", "Bearer invalid"),
	} {
		if err := issueTicket(client, ctx); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s: Unauthenticated を期待しましたが %v でした", name, err)
		}
	}
	if err := issueTicket(client, headers("Authorization", "Bearer "+validToken)); err != nil {
		t.Errorf("有効なトークンで失敗しました: %v", err)
	}
}

func TestInvalidRequestIsRejected(t *testing.T) {
	ts := newTestServer(t, &stubAuthUseCase{}, stubRealtimeServer{})
	client := chatv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)
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
		ts := newTestServer(t, &stubAuthUseCase{}, stubRealtimeServer{err: tt.err})
		client := chatv1connect.NewRealtimeServiceClient(http.DefaultClient, ts.URL)
		if err := issueTicket(client, headers("Authorization", "Bearer "+validToken)); connect.CodeOf(err) != tt.want {
			t.Errorf("%v: %v を期待しましたが %v でした", tt.err, tt.want, err)
		}
	}
}

func TestLoginSetsRefreshTokenCookieForBrowser(t *testing.T) {
	ts := newTestServer(t, &stubAuthUseCase{}, stubRealtimeServer{})
	client := chatv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)

	ctx, callInfo := withHeaders()
	res, err := client.Login(ctx, &chatv1.LoginRequest{Email: "alice@example.com", Password: "password123"})
	if err != nil {
		t.Fatal(err)
	}
	if res.RefreshToken != nil {
		t.Errorf("ブラウザには本文でリフレッシュトークンを返さないはず")
	}
	cookie := callInfo.ResponseHeader().Get("Set-Cookie")
	for _, want := range []string{"__Secure-chat_rt=rt", "Path=/chat.v1.AuthService/", "HttpOnly", "Secure", "SameSite=Strict", "Max-Age="} {
		if !strings.Contains(cookie, want) {
			t.Errorf("Cookie に %q がありません: %s", want, cookie)
		}
	}
	if strings.Contains(cookie, "Domain=") {
		t.Errorf("Cookie に Domain を付けないはず: %s", cookie)
	}
}

func TestNativeClientReceivesRefreshTokenInBody(t *testing.T) {
	uc := &stubAuthUseCase{}
	ts := newTestServer(t, uc, stubRealtimeServer{})
	client := chatv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)

	ctx, callInfo := withHeaders("X-Chat-Client", "native")
	res, err := client.Refresh(ctx, &chatv1.RefreshRequest{RefreshToken: new("body-token")})
	if err != nil {
		t.Fatal(err)
	}
	if res.GetRefreshToken() != "rotated" || callInfo.ResponseHeader().Get("Set-Cookie") != "" {
		t.Errorf("ネイティブアプリには本文で返し Cookie を使わないはず: body=%v cookie=%s", res.RefreshToken, callInfo.ResponseHeader().Get("Set-Cookie"))
	}
	if len(uc.refreshed) != 1 || uc.refreshed[0] != "body-token" {
		t.Errorf("本文のリフレッシュトークンを使っていません: %v", uc.refreshed)
	}
}

func TestRefreshReadsCookieAndChecksOrigin(t *testing.T) {
	uc := &stubAuthUseCase{}
	ts := newTestServer(t, uc, stubRealtimeServer{})
	client := chatv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)

	evil := headers("Cookie", "__Secure-chat_rt=cookie-token", "Origin", "https://evil.example")
	if _, err := client.Refresh(evil, &chatv1.RefreshRequest{}); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Fatalf("許可していないオリジンは拒否するはず: %v", err)
	}

	allowed := headers("Cookie", "__Secure-chat_rt=cookie-token", "Origin", allowedOrigin)
	if _, err := client.Refresh(allowed, &chatv1.RefreshRequest{}); err != nil {
		t.Fatal(err)
	}
	if len(uc.refreshed) != 1 || uc.refreshed[0] != "cookie-token" {
		t.Errorf("Cookie のリフレッシュトークンを使っていません: %v", uc.refreshed)
	}
}

func TestLogoutIsPublicAndClearsCookie(t *testing.T) {
	uc := &stubAuthUseCase{}
	ts := newTestServer(t, uc, stubRealtimeServer{})
	client := chatv1connect.NewAuthServiceClient(http.DefaultClient, ts.URL)

	ctx, callInfo := withHeaders("Cookie", "__Secure-chat_rt=cookie-token", "Authorization", "Bearer "+validToken)
	if _, err := client.Logout(ctx, &chatv1.LogoutRequest{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(callInfo.ResponseHeader().Get("Set-Cookie"), "Max-Age=0") {
		t.Errorf("Cookie を消していません: %s", callInfo.ResponseHeader().Get("Set-Cookie"))
	}
	want := authuc.LogoutInput{RefreshToken: "cookie-token", SessionID: "session-1"}
	if len(uc.loggedOut) != 1 || uc.loggedOut[0] != want {
		t.Errorf("失効させるセッションが期待と異なります: %+v", uc.loggedOut)
	}

	// 期限切れのアクセストークンでもログアウトはできる
	if _, err := client.Logout(context.Background(), &chatv1.LogoutRequest{}); err != nil {
		t.Errorf("トークンなしのログアウトが失敗しました: %v", err)
	}
}

func TestConnectProtocolHeaderIsRequired(t *testing.T) {
	ts := newTestServer(t, &stubAuthUseCase{}, stubRealtimeServer{})
	res, err := http.Post(ts.URL+chatv1connect.RealtimeServiceIssueWebSocketTicketProcedure, "application/json", strings.NewReader(wsRequest))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode == http.StatusOK || res.StatusCode == http.StatusUnauthorized {
		t.Errorf("Connect-Protocol-Version のない呼び出しは拒否するはず: %d", res.StatusCode)
	}
}
