package admin

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type AuditLogQuery struct {
	WorkspaceID string
	RequesterID string
	ActorID     *string
	Actions     []entity.AuditAction
	Since       *time.Time
	Until       *time.Time
}

type ListAuditLogsInput struct {
	AuditLogQuery
	Limit  int
	Offset int
}

type UserSummary struct {
	ID          string
	DisplayName string
	AvatarURL   *string
}

type AuditLogOutput struct {
	entity.AuditLog
	Actor *UserSummary
}

type ListAuditLogsOutput struct {
	Logs       []AuditLogOutput
	TotalCount int
}

type ExportOutput struct {
	Content  string
	FileName string
}

type WorkspaceInput struct {
	WorkspaceID string
	RequesterID string
}

type MemberActionInput struct {
	WorkspaceID  string
	TargetUserID string
	OperatorID   string
}

type MemberOutput struct {
	UserID      string
	Email       string
	DisplayName string
	AvatarURL   *string
	Role        entity.WorkspaceRole
	JoinedAt    time.Time
	SuspendedAt *time.Time
	LastLogin   *entity.Session
	// 2 段階認証は未実装のため常に false
	TwoFactorEnabled bool
	Activity         entity.MemberActivity
}

type PermissionsOutput struct {
	Matrix        entity.PermissionMatrix
	RequesterRole entity.WorkspaceRole
}

type UpdatePermissionInput struct {
	WorkspaceID string
	OperatorID  string
	Role        entity.WorkspaceRole
	Permission  entity.Permission
	Allowed     bool
}
