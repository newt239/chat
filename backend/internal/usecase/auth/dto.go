package auth

import "time"

// Input DTOs

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type RefreshTokenInput struct {
	RefreshToken string
}

type LogoutInput struct {
	UserID       string
	RefreshToken string
}

// Output DTOs

type AuthOutput struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	User         UserInfo  `json:"user"`
}

type UserInfo struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type LogoutOutput struct {
	Success bool `json:"success"`
}
