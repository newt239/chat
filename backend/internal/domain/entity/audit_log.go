package entity

import "time"

type AuditAction string

const (
	AuditActionLogin             AuditAction = "login"
	AuditActionLoginFailed       AuditAction = "login_failed"
	AuditActionMemberRoleChanged AuditAction = "member_role_changed"
	AuditActionMemberSuspended   AuditAction = "member_suspended"
	AuditActionMemberResumed     AuditAction = "member_resumed"
	AuditActionChannelCreated    AuditAction = "channel_created"
	AuditActionChannelDeleted    AuditAction = "channel_deleted"
	AuditActionChannelArchived   AuditAction = "channel_archived"
	AuditActionChannelUnarchived AuditAction = "channel_unarchived"
	AuditActionPermissionChanged AuditAction = "permission_changed"
	AuditActionDataExported      AuditAction = "data_exported"
)

type AuditTargetType string

const (
	AuditTargetUser    AuditTargetType = "user"
	AuditTargetChannel AuditTargetType = "channel"
	AuditTargetRole    AuditTargetType = "role"
	AuditTargetData    AuditTargetType = "data"
)

type AuditLog struct {
	ID          string
	WorkspaceID string
	// ログインの失敗など実行者を特定できない場合は nil
	ActorID     *string
	Action      AuditAction
	TargetType  AuditTargetType
	TargetID    string
	TargetLabel string
	Metadata    map[string]string
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}

type AuditLogFilter struct {
	WorkspaceID string
	ActorID     *string
	Actions     []AuditAction
	Since       *time.Time
	Until       *time.Time
	Limit       int
	Offset      int
}
