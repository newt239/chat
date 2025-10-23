package domain

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

type UserRepository interface {
	FindByID(id string) (*User, error)
	FindByIDs(ids []string) ([]*User, error)
	FindByEmail(email string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id string) error
}
