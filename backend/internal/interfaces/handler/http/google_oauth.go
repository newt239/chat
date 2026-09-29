package http

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

// GoogleOAuthFlow はネイティブアプリの Google ログインで、ブラウザとアプリの間を取り持ちます
type GoogleOAuthFlow interface {
	AuthCodeURL(state, codeChallenge, nonce string) (string, error)
	AppRedirectURL(query url.Values) (string, error)
}

func validLength(v string, minLen, maxLen int) bool {
	return len(v) >= minLen && len(v) <= maxLen
}

// アプリがシステムのブラウザで開き、Google の同意画面へ送る
func googleOAuthStartHandler(flow GoogleOAuthFlow) echo.HandlerFunc {
	return func(c echo.Context) error {
		state := c.QueryParam("state")
		challenge := c.QueryParam("code_challenge")
		nonce := c.QueryParam("nonce")
		if !validLength(state, 16, 256) || !validLength(challenge, 43, 128) || !validLength(nonce, 16, 256) {
			return c.String(http.StatusBadRequest, "invalid request")
		}
		target, err := flow.AuthCodeURL(state, challenge, nonce)
		if err != nil {
			return c.String(http.StatusNotFound, "not found")
		}
		return c.Redirect(http.StatusFound, target)
	}
}

// Google から戻ってきたブラウザをアプリのディープリンクへ送る。コードの交換はアプリが行う
func googleOAuthCallbackHandler(flow GoogleOAuthFlow) echo.HandlerFunc {
	return func(c echo.Context) error {
		target, err := flow.AppRedirectURL(c.QueryParams())
		if err != nil {
			return c.String(http.StatusNotFound, "not found")
		}
		return c.Redirect(http.StatusFound, target)
	}
}
