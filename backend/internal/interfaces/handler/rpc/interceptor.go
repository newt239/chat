package rpc

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/newt239/chat/internal/gen/chat/v1/chatv1connect"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

var publicProcedures = map[string]struct{}{
	chatv1connect.AuthServiceRegisterProcedure: {},
	chatv1connect.AuthServiceLoginProcedure:    {},
	chatv1connect.AuthServiceRefreshProcedure:  {},
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
