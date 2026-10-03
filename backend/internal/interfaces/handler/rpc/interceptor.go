package rpc

import (
	"context"
	"errors"
	"net"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/validate"

	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	"github.com/newt239/chat/internal/usecase/audit"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

// maxRequestBytes は 1 リクエストの本文の上限。ファイルは署名付き URL で直接アップロードするため小さくてよい
const maxRequestBytes = 1 << 20

// HandlerOptions は全サービスに共通の interceptor と制限です
func HandlerOptions(jwtService authuc.JWTService, allowedOrigins []string) []connect.HandlerOption {
	// 外側から順にエラー変換・オリジン確認・操作元の記録・認証・入力検証を適用する
	interceptors := connect.WithInterceptors(
		newErrorInterceptor(),
		newOriginInterceptor(allowedOrigins),
		newClientInfoInterceptor(),
		newAuthInterceptor(jwtService),
		validate.NewInterceptor(),
	)
	// Connect-Protocol-Version ヘッダーを必須にし、フォームなどから単純リクエストで呼ばれる CSRF を防ぐ
	return []connect.HandlerOption{interceptors, connect.WithRequireConnectProtocolHeader(), connect.WithReadMaxBytes(maxRequestBytes)}
}

var publicProcedures = map[string]struct{}{
	chatv1connect.AuthServiceGetAuthConfigProcedure:               {},
	chatv1connect.AuthServiceLoginProcedure:                       {},
	chatv1connect.AuthServiceLoginWithGoogleProcedure:             {},
	chatv1connect.AuthServiceLoginWithGoogleCodeProcedure:         {},
	chatv1connect.AuthServiceSignUpProcedure:                      {},
	chatv1connect.AuthServiceSignUpWithInvitationProcedure:        {},
	chatv1connect.AuthServiceRefreshProcedure:                     {},
	chatv1connect.AuthServiceLogoutProcedure:                      {},
	chatv1connect.InvitationServiceGetInvitationProcedure:         {},
	chatv1connect.WorkspaceServiceGetWorkspaceSignupInfoProcedure: {},
}

// cookieProcedures は Cookie のリフレッシュトークンを使うため、許可していないオリジンからの呼び出しを拒否します
var cookieProcedures = map[string]struct{}{
	chatv1connect.AuthServiceRefreshProcedure: {},
	chatv1connect.AuthServiceLogoutProcedure:  {},
}

type claimsKey struct{}

func claimsFrom(ctx context.Context) authuc.TokenClaims {
	if claims, ok := ctx.Value(claimsKey{}).(*authuc.TokenClaims); ok {
		return *claims
	}
	return authuc.TokenClaims{}
}

func userIDFrom(ctx context.Context) string {
	return claimsFrom(ctx).UserID
}

// newAuthInterceptor は公開 RPC でもアクセストークンがあれば検証し、本人とセッションを context に載せます
func newAuthInterceptor(jwtService authuc.JWTService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			_, public := publicProcedures[req.Spec().Procedure]
			token, ok := strings.CutPrefix(req.Header().Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				if public {
					return next(ctx, req)
				}
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("Authorizationヘッダーが指定されていません"))
			}
			claims, err := jwtService.VerifyToken(token)
			if err != nil {
				if public {
					return next(ctx, req)
				}
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("トークンが無効または期限切れです"))
			}
			return next(context.WithValue(ctx, claimsKey{}, claims), req)
		}
	}
}

// newOriginInterceptor はブラウザから Cookie を使う RPC を呼んだとき、Origin が許可したものかを確かめます。Origin のないネイティブアプリは通す
func newOriginInterceptor(allowedOrigins []string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := cookieProcedures[req.Spec().Procedure]; ok {
				origin := req.Header().Get("Origin")
				if origin != "" && !slices.Contains(allowedOrigins, "*") && !slices.Contains(allowedOrigins, origin) {
					return nil, connect.NewError(connect.CodePermissionDenied, errors.New("許可されていないオリジンです"))
				}
			}
			return next(ctx, req)
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
			return nil, toConnectError(ctx, req.Spec().Procedure, err)
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

// clientIP は信頼するプロキシを経由したときだけ X-Forwarded-For を使うよう、ルーターで書き換えた接続元を返します
func clientIP(req connect.AnyRequest) string {
	host, _, err := net.SplitHostPort(req.Peer().Addr)
	if err != nil {
		return req.Peer().Addr
	}
	return host
}
