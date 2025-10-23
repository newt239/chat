package domain

import "time"

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

type SessionRepository interface {
	FindByID(id string) (*Session, error)
	FindActiveByUserID(userID string) ([]*Session, error)
	Create(session *Session) error
	Revoke(id string) error
	RevokeAllByUserID(userID string) error
	DeleteExpired() error
}
