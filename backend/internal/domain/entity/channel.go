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

type ChannelRole string

const (
	ChannelRoleMember ChannelRole = "member"
	ChannelRoleAdmin  ChannelRole = "admin"
)

type ChannelMember struct {
	ChannelID string
	UserID    string
	Role      ChannelRole
	JoinedAt  time.Time
}
