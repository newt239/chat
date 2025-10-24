package entity

import "time"

type MessageBookmark struct {
	UserID    string
	MessageID string
	CreatedAt time.Time
}
