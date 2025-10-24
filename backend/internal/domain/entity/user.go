package entity

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	AvatarURL    *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
