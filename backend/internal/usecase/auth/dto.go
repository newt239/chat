package auth

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// Settings は設定ファイルから渡す認証の挙動です
type Settings struct {
	AccessTokenTTL      time.Duration
	RefreshTokenTTL     time.Duration
	PasswordAuthEnabled bool
}

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

// LogoutInput はブラウザなら Cookie のリフレッシュトークン、ネイティブアプリならアクセストークンのセッション ID で失効させるセッションを指します
type LogoutInput struct {
	RefreshToken string
	SessionID    string
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

type AuthOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         *entity.User
}
