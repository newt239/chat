package auth

import (
	"context"

	"google.golang.org/api/idtoken"

	domerr "github.com/newt239/chat/internal/domain/errors"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

var googleIssuers = map[string]bool{"accounts.google.com": true, "https://accounts.google.com": true}

// GoogleVerifier は ClientID が空なら常に ErrGoogleAuthDisabled を返します
type GoogleVerifier struct {
	ClientID string
}

func (v GoogleVerifier) Verify(ctx context.Context, token string) (*authuc.GoogleIdentity, error) {
	if v.ClientID == "" {
		return nil, domerr.ErrGoogleAuthDisabled
	}
	payload, err := idtoken.Validate(ctx, token, v.ClientID)
	if err != nil {
		return nil, domerr.ErrInvalidToken
	}
	if !googleIssuers[payload.Issuer] {
		return nil, domerr.ErrInvalidToken
	}
	claims := payload.Claims
	email, _ := claims["email"].(string)
	verified, _ := claims["email_verified"].(bool)
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)
	nonce, _ := claims["nonce"].(string)
	return &authuc.GoogleIdentity{Sub: payload.Subject, Email: email, EmailVerified: verified, Name: name, Picture: picture, Nonce: nonce}, nil
}
