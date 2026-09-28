package webhook

import (
	"time"

	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type ListInput struct {
	ChannelID string
	UserID    string
}

type CreateInput struct {
	ChannelID string
	UserID    string
	Name      string
	AvatarURL *string
}

type UpdateInput struct {
	WebhookID string
	UserID    string
	Name      string
	AvatarURL *string
}

type TargetInput struct {
	WebhookID string
	UserID    string
}

// PostInput は外部から届いた投稿です。Username と AvatarURL はこの投稿の表示だけを上書きします
type PostInput struct {
	WebhookID string
	Token     string
	Text      string
	Username  *string
	AvatarURL *string
}

type Output struct {
	ID         string
	ChannelID  string
	Name       string
	AvatarURL  *string
	CreatedBy  messageuc.UserInfo
	CreatedAt  time.Time
	LastUsedAt *time.Time
	CanManage  bool
}

type CreateOutput struct {
	Webhook Output
	Token   string
}
