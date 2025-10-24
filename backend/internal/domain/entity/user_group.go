package entity

import "time"

type UserGroup struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserGroupMember struct {
	GroupID  string
	UserID   string
	JoinedAt time.Time
}
