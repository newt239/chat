package domain

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

type WorkspaceRepository interface {
	FindByID(id string) (*Workspace, error)
	FindByUserID(userID string) ([]*Workspace, error)
	Create(workspace *Workspace) error
	Update(workspace *Workspace) error
	Delete(id string) error
	AddMember(member *WorkspaceMember) error
	UpdateMemberRole(workspaceID, userID string, role WorkspaceRole) error
	RemoveMember(workspaceID, userID string) error
	FindMembersByWorkspaceID(workspaceID string) ([]*WorkspaceMember, error)
	FindMember(workspaceID, userID string) (*WorkspaceMember, error)
}
