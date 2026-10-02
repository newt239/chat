package rpc

import (
	"context"

	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type claimsKey struct{}

func withClaims(ctx context.Context, claims *authuc.TokenClaims) context.Context {
	return context.WithValue(ctx, claimsKey{}, claims)
}

func claimsFrom(ctx context.Context) authuc.TokenClaims {
	if claims, ok := ctx.Value(claimsKey{}).(*authuc.TokenClaims); ok {
		return *claims
	}
	return authuc.TokenClaims{}
}

// userIDFrom は認証 interceptor が設定したユーザー ID を返します
func userIDFrom(ctx context.Context) string {
	return claimsFrom(ctx).UserID
}
