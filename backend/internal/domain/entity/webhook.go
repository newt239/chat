package entity

import (
	"crypto/subtle"
	"time"
)

// Webhook はチャンネルへの着信 Webhook です。トークンは発行時にだけ平文で返し、ハッシュで保存します
type Webhook struct {
	ID         string
	ChannelID  string
	Name       string
	AvatarURL  *string
	TokenHash  string
	BotUserID  string
	CreatedBy  string
	LastUsedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (w *Webhook) VerifyToken(token string) bool {
	return subtle.ConstantTimeCompare([]byte(HashSecretToken(token)), []byte(w.TokenHash)) == 1
}
