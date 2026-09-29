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

type LoginWithGoogleInput struct {
	IDToken string
	// 参加リンクから来た場合のワークスペース。招待がなくても登録を許可していればアカウントを作る
	WorkspaceID *string
}

type LoginWithGoogleCodeInput struct {
	Code         string
	CodeVerifier string
	Nonce        string
	WorkspaceID  *string
}

type SignUpInput struct {
	WorkspaceID string
	Email       string
	DisplayName string
	Password    string
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
	// 認可コードフローで認可リクエストに付けた値。ID トークンだけを受け取るフローでは空
	Nonce string
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
