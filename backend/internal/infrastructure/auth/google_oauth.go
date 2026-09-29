package auth

import (
	"context"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
)

// GoogleOAuth はネイティブアプリの Google ログインで、ブラウザでの認可コードフロー（PKCE）を仲介します
type GoogleOAuth struct {
	config         oauth2.Config
	appRedirectURL string
}

// NewGoogleOAuth は値が一つでも空なら、常に ErrGoogleAuthDisabled を返すものを作ります
func NewGoogleOAuth(clientID, clientSecret, redirectURL, appRedirectURL string) *GoogleOAuth {
	return &GoogleOAuth{
		config: oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     endpoints.Google,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
		},
		appRedirectURL: appRedirectURL,
	}
}

func (g *GoogleOAuth) enabled() bool {
	return g.config.ClientID != "" && g.config.ClientSecret != "" && g.config.RedirectURL != "" && g.appRedirectURL != ""
}

// AuthCodeURL は Google の同意画面の URL を返します。code_verifier はアプリだけが持つ
func (g *GoogleOAuth) AuthCodeURL(state, codeChallenge, nonce string) (string, error) {
	if !g.enabled() {
		return "", domainerrors.ErrGoogleAuthDisabled
	}
	return g.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	), nil
}

// AppRedirectURL は Google から戻ってきたクエリをアプリのディープリンクに載せ替えます
func (g *GoogleOAuth) AppRedirectURL(query url.Values) (string, error) {
	if !g.enabled() {
		return "", domainerrors.ErrGoogleAuthDisabled
	}
	u, err := url.Parse(g.appRedirectURL)
	if err != nil {
		return "", err
	}
	forwarded := url.Values{}
	for _, key := range []string{"code", "state", "error"} {
		if v := query.Get(key); v != "" {
			forwarded.Set(key, v)
		}
	}
	u.RawQuery = forwarded.Encode()
	return u.String(), nil
}

func (g *GoogleOAuth) Exchange(ctx context.Context, code, codeVerifier string) (string, error) {
	if !g.enabled() {
		return "", domainerrors.ErrGoogleAuthDisabled
	}
	token, err := g.config.Exchange(ctx, code, oauth2.VerifierOption(codeVerifier))
	if err != nil {
		return "", domainerrors.ErrInvalidToken
	}
	idToken, ok := token.Extra("id_token").(string)
	if !ok || idToken == "" {
		return "", domainerrors.ErrInvalidToken
	}
	return idToken, nil
}
