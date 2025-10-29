package entity

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	AvatarURL    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
