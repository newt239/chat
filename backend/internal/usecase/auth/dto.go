package auth

import "time"

// Settings は設定ファイルから渡す認証の挙動です
type Settings struct {
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	PasswordAuthEnabled bool
}

// Input DTOs

type LoginInput struct {
	Email    string
	Password string
}

type SignUpWithInvitationInput struct {
	Token       string
	DisplayName string
	Password    string
}

type RefreshTokenInput struct {
	RefreshToken string
}

type LogoutInput struct {
	UserID string
}

// GoogleIdentity は検証済みの Google ID トークンから取り出した本人情報です
type GoogleIdentity struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
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
