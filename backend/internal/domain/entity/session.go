package entity

import "time"

type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	IPAddress        string
	UserAgent        string
	CreatedAt        time.Time
}
