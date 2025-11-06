package entity

import (
    "errors"
    "regexp"
    "time"
)

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
    IsPublic    bool
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

func (m *WorkspaceMember) CanCreateChannel() bool {
	if m == nil {
		return false
	}

	return m.Role == WorkspaceRoleOwner || m.Role == WorkspaceRoleAdmin
}

// ValidateWorkspaceSlug validates the workspace slug format and length.
func ValidateWorkspaceSlug(slug string) error {
    if len(slug) < 3 || len(slug) > 12 {
        return errors.New("ワークスペースIDは3〜12文字である必要があります")
    }

    matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`, slug)
    if !matched {
        return errors.New("ワークスペースIDは英小文字、数字、ハイフンのみ使用できます")
    }

    return nil
}
