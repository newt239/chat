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
	Limit     int
	PageToken string
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
	Logs          []AuditLogOutput
	NextPageToken string
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
	Activity    entity.MemberActivity
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
