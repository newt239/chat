package auth

import "time"

type TokenClaims struct {
	UserID string
	Email  string
}

type JWTService interface {
	GenerateToken(userID string, duration time.Duration) (string, error)
	VerifyToken(token string) (*TokenClaims, error)
}

type PasswordService interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) error
}
