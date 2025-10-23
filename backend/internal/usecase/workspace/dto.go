package workspace

import "time"

// Input DTOs

type CreateWorkspaceInput struct {
	Name        string
	Description *string
	CreatedBy   string
}

type UpdateWorkspaceInput struct {
	ID          string
	Name        *string
	Description *string
	IconURL     *string
	UserID      string // For authorization check
}

type DeleteWorkspaceInput struct {
	ID     string
	UserID string // For authorization check
}

type GetWorkspaceInput struct {
	ID     string
	UserID string // For authorization check
}

type AddMemberInput struct {
	WorkspaceID string
	UserID      string
	InviterID   string // User performing the action
	Role        string
}

type UpdateMemberRoleInput struct {
	WorkspaceID string
	UserID      string
	UpdaterID   string // User performing the action
	Role        string
}

type RemoveMemberInput struct {
	WorkspaceID string
	UserID      string
	RemoverID   string // User performing the action
}

type ListMembersInput struct {
	WorkspaceID string
	RequesterID string // For authorization check
}

// Output DTOs

// WorkspaceOutput represents a workspace in the response
type WorkspaceOutput struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IconURL     *string   `json:"iconUrl"`
	Role        string    `json:"role"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// GetWorkspacesOutput represents the output of getting workspaces
type GetWorkspacesOutput struct {
	Workspaces []WorkspaceOutput `json:"workspaces"`
}

type GetWorkspaceOutput struct {
	Workspace WorkspaceOutput `json:"workspace"`
}

type CreateWorkspaceOutput struct {
	Workspace WorkspaceOutput `json:"workspace"`
}

type UpdateWorkspaceOutput struct {
	Workspace WorkspaceOutput `json:"workspace"`
}

type DeleteWorkspaceOutput struct {
	Success bool `json:"success"`
}

type MemberInfo struct {
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl,omitempty"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
}

type ListMembersOutput struct {
	Members []MemberInfo `json:"members"`
}

type MemberActionOutput struct {
	Success bool `json:"success"`
}
