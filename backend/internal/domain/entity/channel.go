package entity

import "time"

type Channel struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	IsPrivate   bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChannelMember struct {
	ChannelID string
	UserID    string
	JoinedAt  time.Time
}
