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
	SuspendedAt *time.Time
	// FindMembershipsByUserID のときだけ設定される
	Workspace *Workspace
}

// IsAdmin はワークスペースの管理画面を操作できるロールかを返します
func (m *WorkspaceMember) IsAdmin() bool {
	return m.Role == WorkspaceRoleOwner || m.Role == WorkspaceRoleAdmin
}

// MemberActivity は管理画面に出すメンバーの直近の投稿数と添付の合計サイズです
type MemberActivity struct {
	MessageCount  int
	StorageBytes  int64
	LastMessageAt *time.Time
}
