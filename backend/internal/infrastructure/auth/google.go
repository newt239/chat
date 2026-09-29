package auth

import (
	"context"

	"google.golang.org/api/idtoken"

	domainerrors "github.com/newt239/chat/internal/domain/errors"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

var googleIssuers = map[string]bool{"accounts.google.com": true, "https://accounts.google.com": true}

type googleVerifier struct {
	clientID string
}

// NewGoogleVerifier は clientID が空なら常に ErrGoogleAuthDisabled を返す検証器を作ります
func NewGoogleVerifier(clientID string) authuc.GoogleVerifier {
	return &googleVerifier{clientID: clientID}
}

func (v *googleVerifier) Verify(ctx context.Context, token string) (*authuc.GoogleIdentity, error) {
	if v.clientID == "" {
		return nil, domainerrors.ErrGoogleAuthDisabled
	}
	payload, err := idtoken.Validate(ctx, token, v.clientID)
	if err != nil {
		return nil, domainerrors.ErrInvalidToken
	}
	if !googleIssuers[payload.Issuer] {
		return nil, domainerrors.ErrInvalidToken
	}
	return identityFromClaims(payload.Subject, payload.Claims), nil
}

func identityFromClaims(sub string, claims map[string]any) *authuc.GoogleIdentity {
	email, _ := claims["email"].(string)
	verified, _ := claims["email_verified"].(bool)
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)
	return &authuc.GoogleIdentity{Sub: sub, Email: email, EmailVerified: verified, Name: name, Picture: picture}
}
