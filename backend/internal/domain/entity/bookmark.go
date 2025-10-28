package entity

import "time"

type MessageBookmark struct {
	UserID    string
	MessageID string
	Message   *Message
	CreatedAt time.Time
}
