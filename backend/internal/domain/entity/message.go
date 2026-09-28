package entity

import "time"

type Message struct {
	ID        string
	ChannelID string
	UserID    string
	ParentID  *string
	Body      string
	CreatedAt time.Time
	EditedAt  *time.Time
	DeletedAt *time.Time
	DeletedBy *string
	// Webhook が投稿ごとに指定した表示名とアイコン
	SenderName      *string
	SenderAvatarURL *string
}

type MessageReaction struct {
	MessageID string
	UserID    string
	Emoji     string
	CreatedAt time.Time
}
