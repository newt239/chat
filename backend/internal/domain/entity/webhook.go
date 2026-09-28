package entity

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
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

// NewWebhookToken は URL に埋め込むトークンとその保存用ハッシュを生成します
func NewWebhookToken() (token string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, HashWebhookToken(token), nil
}

// HashWebhookToken は十分な長さの乱数トークンを前提に SHA-256 でハッシュします
func HashWebhookToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (w *Webhook) VerifyToken(token string) bool {
	return subtle.ConstantTimeCompare([]byte(HashWebhookToken(token)), []byte(w.TokenHash)) == 1
}
