package entity

import (
	"fmt"
	"regexp"
	"time"

	domerr "github.com/newt239/chat/internal/domain/errors"
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
	// 招待なしで参加リンクからアカウントを作れるか
	SignupEnabled      bool
	EmailSignupEnabled bool
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type WorkspaceMember struct {
	WorkspaceID string
	UserID      string
	Role        WorkspaceRole
	JoinedAt    time.Time
	SuspendedAt *time.Time
}

// IsAdmin はワークスペースの管理画面を操作できるロールかを返します
func (m *WorkspaceMember) IsAdmin() bool {
	if m == nil {
		return false
	}

	return m.Role == WorkspaceRoleOwner || m.Role == WorkspaceRoleAdmin
}

var (
	ErrWorkspaceSlugInvalid = fmt.Errorf("%w: ワークスペースIDは英小文字・数字・ハイフンの 3〜12 文字で指定してください", domerr.ErrValidation)
	workspaceSlugPattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
)

// ValidateWorkspaceSlug はワークスペース ID の長さと使える文字を確かめます
func ValidateWorkspaceSlug(slug string) error {
	if len(slug) < 3 || len(slug) > 12 || !workspaceSlugPattern.MatchString(slug) {
		return ErrWorkspaceSlugInvalid
	}
	return nil
}
