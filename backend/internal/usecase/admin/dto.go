package admin

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/usecase/message"
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

type AuditLogOutput struct {
	entity.AuditLog
	Actor *message.UserInfo
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

type UpdateMemberRoleInput struct {
	MemberActionInput
	Role entity.WorkspaceRole
}

type MemberOutput struct {
	*entity.WorkspaceMember
	User      *entity.User
	LastLogin *entity.Session
	Activity  entity.MemberActivity
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
