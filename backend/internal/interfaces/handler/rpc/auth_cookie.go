package rpc

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
)

const (
	refreshTokenCookieName = "__Secure-chat_rt"
	refreshTokenCookiePath = "/chat.v1.AuthService/"
	clientHeader           = "X-Chat-Client"
	nativeClient           = "native"
)

// isNativeClient は Cookie を扱えないネイティブアプリからの呼び出しかを返します。リフレッシュトークンは本文で受け渡す
func isNativeClient(ctx context.Context) bool {
	info, ok := connect.CallInfoForHandlerContext(ctx)
	return ok && info.RequestHeader().Get(clientHeader) == nativeClient
}

func refreshTokenFromCookie(ctx context.Context) string {
	info, ok := connect.CallInfoForHandlerContext(ctx)
	if !ok {
		return ""
	}
	cookie, err := (&http.Request{Header: info.RequestHeader()}).Cookie(refreshTokenCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setRefreshTokenCookie(ctx context.Context, token string, expiresAt time.Time) {
	writeRefreshTokenCookie(ctx, token, int(time.Until(expiresAt).Seconds()))
}

func clearRefreshTokenCookie(ctx context.Context) {
	writeRefreshTokenCookie(ctx, "", -1)
}

func writeRefreshTokenCookie(ctx context.Context, value string, maxAge int) {
	info, ok := connect.CallInfoForHandlerContext(ctx)
	if !ok {
		return
	}
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    value,
		Path:     refreshTokenCookiePath,
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	info.ResponseHeader().Add("Set-Cookie", cookie.String())
}
