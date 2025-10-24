package entity

import "time"

type WorkspaceRole string

const (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
	WorkspaceRoleGuest  WorkspaceRole = "guest"
)

type Workspace struct {
	ID          string
	Name        string
	Description *string
	IconURL     *string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type WorkspaceMember struct {
	WorkspaceID string
	UserID      string
	Role        WorkspaceRole
	JoinedAt    time.Time
}
