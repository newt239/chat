package entity

import "time"

type AuditAction string

const (
	AuditActionLogin              AuditAction = "login"
	AuditActionLoginFailed        AuditAction = "login_failed"
	AuditActionMemberRoleChanged  AuditAction = "member_role_changed"
	AuditActionMemberSuspended    AuditAction = "member_suspended"
	AuditActionMemberResumed      AuditAction = "member_resumed"
	AuditActionChannelCreated     AuditAction = "channel_created"
	AuditActionPermissionChanged  AuditAction = "permission_changed"
	AuditActionAuditLogExported   AuditAction = "audit_log_exported"
	AuditActionAppCreated         AuditAction = "app_created"
	AuditActionAppDeleted         AuditAction = "app_deleted"
	AuditActionCustomEmojiCreated AuditAction = "custom_emoji_created"
	AuditActionCustomEmojiDeleted AuditAction = "custom_emoji_deleted"
)

type AuditTargetType string

const (
	AuditTargetUser        AuditTargetType = "user"
	AuditTargetChannel     AuditTargetType = "channel"
	AuditTargetRole        AuditTargetType = "role"
	AuditTargetApp         AuditTargetType = "app"
	AuditTargetCustomEmoji AuditTargetType = "custom_emoji"
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

// AuditLogFilter の期間は Since 以上 Until 未満です
type AuditLogFilter struct {
	WorkspaceID string
	ActorID     *string
	Actions     []AuditAction
	Since       *time.Time
	Until       *time.Time
}
