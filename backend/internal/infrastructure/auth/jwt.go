package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authuc "github.com/newt239/chat/internal/usecase/auth"
)

var (
	ErrInvalidToken = errors.New("トークンが無効です")
	ErrExpiredToken = errors.New("トークンの有効期限が切れています")
)

type Claims struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secret string
}

func NewJWTService(secret string) authuc.JWTService {
	return &jwtService{
		secret: secret,
	}
}

func (s *jwtService) GenerateToken(claims authuc.TokenClaims, duration time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})
	return token.SignedString([]byte(s.secret))
}

func (s *jwtService) VerifyToken(tokenString string) (*authuc.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &authuc.TokenClaims{UserID: claims.UserID, SessionID: claims.SessionID}, nil
}
