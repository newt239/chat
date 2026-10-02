package workspace

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// Input DTOs

type CreateWorkspaceInput struct {
	ID          string
	Name        string
	Description *string
	IconURL     *string
	IsPublic    bool
	CreatedBy   string
}

type UpdateWorkspaceInput struct {
	ID                 string
	Name               *string
	Description        *string
	IconURL            *string
	IsPublic           *bool
	SignupEnabled      *bool
	EmailSignupEnabled *bool
	UserID             string // For authorization check
}

type DeleteWorkspaceInput struct {
	ID     string
	UserID string // For authorization check
}

type GetWorkspaceInput struct {
	ID     string
	UserID string // For authorization check
}

type UpdateMemberRoleInput struct {
	WorkspaceID string
	UserID      string
	UpdaterID   string // User performing the action
	Role        entity.WorkspaceRole
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
	ID                 string               `json:"id"`
	Name               string               `json:"name"`
	Description        *string              `json:"description"`
	IconURL            *string              `json:"iconUrl"`
	IsPublic           bool                 `json:"isPublic"`
	SignupEnabled      bool                 `json:"signupEnabled"`
	EmailSignupEnabled bool                 `json:"emailSignupEnabled"`
	Role               entity.WorkspaceRole `json:"role"`
	CreatedBy          string               `json:"createdBy"`
	CreatedAt          time.Time            `json:"createdAt"`
	UpdatedAt          time.Time            `json:"updatedAt"`
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
	UserID      string               `json:"userId"`
	Email       string               `json:"email"`
	DisplayName string               `json:"displayName"`
	AvatarURL   *string              `json:"avatarUrl,omitempty"`
	Bio         *string              `json:"bio,omitempty"`
	Role        entity.WorkspaceRole `json:"role"`
	JoinedAt    time.Time            `json:"joinedAt"`
	SuspendedAt *time.Time           `json:"suspendedAt,omitempty"`
	// 取得したユーザーだけに見えるニックネーム
	Nickname *string  `json:"nickname,omitempty"`
	Timezone string   `json:"timezone"`
	Links    []string `json:"links"`
}

type ListMembersOutput struct {
	Members []MemberInfo `json:"members"`
}

type MemberActionOutput struct {
	Success bool `json:"success"`
}

// Public workspaces
type PublicWorkspaceItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IconURL     *string   `json:"iconUrl"`
	MemberCount int       `json:"memberCount"`
	IsJoined    bool      `json:"isJoined"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ListPublicWorkspacesOutput struct {
	Workspaces []PublicWorkspaceItem `json:"workspaces"`
}

type JoinPublicWorkspaceInput struct {
	WorkspaceID string
	UserID      string
}

type SignupInfoOutput struct {
	ID                 string
	Name               string
	IconURL            *string
	EmailSignupEnabled bool
}
