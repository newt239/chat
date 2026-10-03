package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	domerr "github.com/newt239/chat/internal/domain/errors"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (s *JWTService) GenerateToken(c authuc.TokenClaims, duration time.Duration) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID:    c.UserID,
		SessionID: c.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString(s.secret)
}

func (s *JWTService) VerifyToken(tokenString string) (*authuc.TokenClaims, error) {
	var c claims
	token, err := jwt.ParseWithClaims(tokenString, &c, func(*jwt.Token) (any, error) { return s.secret, nil }, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, domerr.ErrInvalidToken
	}
	return &authuc.TokenClaims{UserID: c.UserID, SessionID: c.SessionID}, nil
}
