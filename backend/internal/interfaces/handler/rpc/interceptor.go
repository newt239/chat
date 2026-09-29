package rpc

import (
	"context"
	"errors"
	"net"
	"strings"

	"connectrpc.com/connect"

	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	"github.com/newt239/chat/internal/usecase/audit"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

var publicProcedures = map[string]struct{}{
	chatv1connect.AuthServiceGetAuthConfigProcedure:               {},
	chatv1connect.AuthServiceLoginProcedure:                       {},
	chatv1connect.AuthServiceLoginWithGoogleProcedure:             {},
	chatv1connect.AuthServiceLoginWithGoogleCodeProcedure:         {},
	chatv1connect.AuthServiceSignUpProcedure:                      {},
	chatv1connect.AuthServiceSignUpWithInvitationProcedure:        {},
	chatv1connect.AuthServiceRefreshProcedure:                     {},
	chatv1connect.InvitationServiceGetInvitationProcedure:         {},
	chatv1connect.WorkspaceServiceGetWorkspaceSignupInfoProcedure: {},
}

func newAuthInterceptor(jwtService authuc.JWTService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := publicProcedures[req.Spec().Procedure]; ok {
				return next(ctx, req)
			}
			token, ok := strings.CutPrefix(req.Header().Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Authorizationヘッダーが指定されていません"))
			}
			claims, err := jwtService.VerifyToken(token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("トークンが無効または期限切れです"))
			}
			return next(withUserID(ctx, claims.UserID), req)
		}
	}
}

// newErrorInterceptor はハンドラが返したユースケースのエラーを Connect のエラーコードに変換します
func newErrorInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			res, err := next(ctx, req)
			if err == nil {
				return res, nil
			}
			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				return nil, err
			}
			return nil, toConnectError(req.Spec().Procedure, err)
		}
	}
}

// newClientInfoInterceptor は監査ログやセッションに残す操作元の IP アドレスと User-Agent を context に載せます
func newClientInfoInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			info := audit.ClientInfo{IPAddress: clientIP(req), UserAgent: req.Header().Get("User-Agent")}
			return next(audit.WithClientInfo(ctx, info), req)
		}
	}
}

// clientIP はロードバランサーを経由する前提で X-Forwarded-For の先頭を優先します
func clientIP(req connect.AnyRequest) string {
	if forwarded := req.Header().Get("X-Forwarded-For"); forwarded != "" {
		first, _, _ := strings.Cut(forwarded, ",")
		return strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(req.Peer().Addr)
	if err != nil {
		return req.Peer().Addr
	}
	return host
}
