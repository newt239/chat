package entity

import "time"

type ScheduledMessageStatus string

const (
	ScheduledMessageScheduled ScheduledMessageStatus = "scheduled"
	// ScheduledMessageSending はワーカーが取り出して投稿している最中
	ScheduledMessageSending ScheduledMessageStatus = "sending"
	ScheduledMessageSent    ScheduledMessageStatus = "sent"
	ScheduledMessageFailed  ScheduledMessageStatus = "failed"
)

// ScheduledMessage は ScheduledAt にサーバーが代わりに投稿するメッセージです
type ScheduledMessage struct {
	ID            string
	UserID        string
	ChannelID     string
	ParentID      *string
	Body          string
	AttachmentIDs []string
	Location      *MessageLocation
	ScheduledAt   time.Time
	Status        ScheduledMessageStatus
	SentMessageID *string
	FailureReason *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsEditable は本文・日時の変更や今すぐ送信ができる状態かを返します
func (m *ScheduledMessage) IsEditable() bool {
	return m.Status == ScheduledMessageScheduled || m.Status == ScheduledMessageFailed
}
