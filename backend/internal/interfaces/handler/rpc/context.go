package rpc

import "context"

type userIDKey struct{}

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// userIDFrom は認証 interceptor が設定したユーザー ID を返します
func userIDFrom(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey{}).(string)
	return userID
}
