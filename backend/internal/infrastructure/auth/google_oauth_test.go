package auth

import (
	"errors"
	"net/url"
	"testing"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
)

func TestGoogleOAuthAuthCodeURL(t *testing.T) {
	g := NewGoogleOAuth("client", "secret", "https://api.example.com/oauth/google/callback", "dev.newt239.chat://auth/callback")

	raw, err := g.AuthCodeURL("state-1", "challenge", "nonce-1")
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	u, _ := url.Parse(raw)
	q := u.Query()
	if u.Host != "accounts.google.com" || q.Get("client_id") != "client" || q.Get("redirect_uri") != "https://api.example.com/oauth/google/callback" {
		t.Errorf("認可 URL が期待と異なります: %s", raw)
	}
	if q.Get("state") != "state-1" || q.Get("code_challenge") != "challenge" || q.Get("code_challenge_method") != "S256" || q.Get("nonce") != "nonce-1" {
		t.Errorf("PKCE と nonce のパラメータが期待と異なります: %s", raw)
	}
}

func TestGoogleOAuthAppRedirectURLForwardsOnlyKnownParams(t *testing.T) {
	g := NewGoogleOAuth("client", "secret", "https://api.example.com/oauth/google/callback", "dev.newt239.chat://auth/callback")

	got, err := g.AppRedirectURL(url.Values{"code": {"c"}, "state": {"s"}, "scope": {"email"}, "authuser": {"0"}})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if got != "dev.newt239.chat://auth/callback?code=c&state=s" {
		t.Errorf("アプリへのリダイレクト先が期待と異なります: %s", got)
	}
}

func TestGoogleOAuthDisabledWithoutSecret(t *testing.T) {
	g := NewGoogleOAuth("client", "", "https://api.example.com/oauth/google/callback", "dev.newt239.chat://auth/callback")

	if _, err := g.AuthCodeURL("state-1", "challenge", "nonce-1"); !errors.Is(err, domainerrors.ErrGoogleAuthDisabled) {
		t.Errorf("client secret がないのに認可 URL を返しました: %v", err)
	}
}
