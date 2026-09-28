package user

import "github.com/newt239/chat/internal/domain/entity"

type UpdateMeInput struct {
	UserID      string
	DisplayName *string
	Bio         *string
	AvatarURL   *string
}

// MeOutput は自分のプロフィールの出力です
type MeOutput struct {
	ID          string                 `json:"id"`
	Email       string                 `json:"email"`
	DisplayName string                 `json:"displayName"`
	Bio         *string                `json:"bio"`
	AvatarURL   *string                `json:"avatarUrl"`
	Preferences entity.UserPreferences `json:"preferences"`
}

type UpdatePreferencesInput struct {
	UserID      string
	Preferences entity.UserPreferences
}

type UpdatePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}
