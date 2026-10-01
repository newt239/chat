package entity

import (
	"fmt"
	"time"

	domerr "github.com/newt239/chat/internal/domain/errors"
)

// ErrInvalidAuditLogPageToken は保存先が発行していない、または壊れたページトークンです
var ErrInvalidAuditLogPageToken = fmt.Errorf("%w: ページトークンが不正です", domerr.ErrValidation)

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
	AuditActionAuditLogExported  AuditAction = "audit_log_exported"
	// アプリに統合する前の着信 Webhook の記録
	AuditActionWebhookCreated     AuditAction = "webhook_created"
	AuditActionWebhookDeleted     AuditAction = "webhook_deleted"
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
	AuditTargetWebhook     AuditTargetType = "webhook"
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

type AuditLogFilter struct {
	WorkspaceID string
	ActorID     *string
	Actions     []AuditAction
	// Since 以上 Until 未満
	Since *time.Time
	Until *time.Time
	Limit int
	// 前ページの AuditLogPage.NextPageToken。形式は保存先ごとに異なる
	PageToken string
}

type AuditLogPage struct {
	// 新しい順
	Logs []*AuditLog
	// 続きがない場合は空
	NextPageToken string
}
