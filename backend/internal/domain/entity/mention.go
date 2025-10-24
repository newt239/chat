package entity

import "time"

type MessageUserMention struct {
	MessageID string
	UserID    string
	CreatedAt time.Time
}

type MessageGroupMention struct {
	MessageID string
	GroupID   string
	CreatedAt time.Time
}
