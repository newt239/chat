package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
)

var auditActions = map[entity.AuditAction]chatv1.AuditAction{
	entity.AuditActionLogin:             chatv1.AuditAction_AUDIT_ACTION_LOGIN,
	entity.AuditActionLoginFailed:       chatv1.AuditAction_AUDIT_ACTION_LOGIN_FAILED,
	entity.AuditActionMemberRoleChanged: chatv1.AuditAction_AUDIT_ACTION_MEMBER_ROLE_CHANGED,
	entity.AuditActionMemberSuspended:   chatv1.AuditAction_AUDIT_ACTION_MEMBER_SUSPENDED,
	entity.AuditActionMemberResumed:     chatv1.AuditAction_AUDIT_ACTION_MEMBER_RESUMED,
	entity.AuditActionChannelCreated:    chatv1.AuditAction_AUDIT_ACTION_CHANNEL_CREATED,
	entity.AuditActionChannelDeleted:    chatv1.AuditAction_AUDIT_ACTION_CHANNEL_DELETED,
	entity.AuditActionChannelArchived:   chatv1.AuditAction_AUDIT_ACTION_CHANNEL_ARCHIVED,
	entity.AuditActionChannelUnarchived: chatv1.AuditAction_AUDIT_ACTION_CHANNEL_UNARCHIVED,
	entity.AuditActionPermissionChanged: chatv1.AuditAction_AUDIT_ACTION_PERMISSION_CHANGED,
	entity.AuditActionAuditLogExported:  chatv1.AuditAction_AUDIT_ACTION_AUDIT_LOG_EXPORTED,
	entity.AuditActionWebhookCreated:    chatv1.AuditAction_AUDIT_ACTION_WEBHOOK_CREATED,
	entity.AuditActionWebhookDeleted:    chatv1.AuditAction_AUDIT_ACTION_WEBHOOK_DELETED,
}

var permissions = map[entity.Permission]chatv1.Permission{
	entity.PermissionCreatePublicChannel:  chatv1.Permission_PERMISSION_CREATE_PUBLIC_CHANNEL,
	entity.PermissionCreatePrivateChannel: chatv1.Permission_PERMISSION_CREATE_PRIVATE_CHANNEL,
	entity.PermissionInviteMembers:        chatv1.Permission_PERMISSION_INVITE_MEMBERS,
	entity.PermissionEditChannelLinks:     chatv1.Permission_PERMISSION_EDIT_CHANNEL_LINKS,
	entity.PermissionPinMessages:          chatv1.Permission_PERMISSION_PIN_MESSAGES,
	entity.PermissionDeleteOthersMessages: chatv1.Permission_PERMISSION_DELETE_OTHERS_MESSAGES,
}

// AuditActionNames はリクエストの操作の種類をユースケースが扱う値に変換します
func AuditActionNames(actions []chatv1.AuditAction) []entity.AuditAction {
	names := make([]entity.AuditAction, 0, len(actions))
	for _, a := range actions {
		for name, v := range auditActions {
			if v == a {
				names = append(names, name)
			}
		}
	}
	return names
}

// PermissionName はリクエストの権限をユースケースが扱う値に変換します
func PermissionName(p chatv1.Permission) entity.Permission {
	for name, v := range permissions {
		if v == p {
			return name
		}
	}
	return ""
}

func AuditLog(l adminuc.AuditLogOutput) *chatv1.AuditLog {
	var actor *chatv1.UserSummary
	if l.Actor != nil {
		actor = &chatv1.UserSummary{Id: l.Actor.ID, DisplayName: l.Actor.DisplayName, AvatarUrl: l.Actor.AvatarURL}
	}
	return &chatv1.AuditLog{
		Id:          l.ID,
		Actor:       actor,
		Action:      auditActions[l.Action],
		TargetType:  string(l.TargetType),
		TargetId:    l.TargetID,
		TargetLabel: l.TargetLabel,
		Metadata:    l.Metadata,
		IpAddress:   l.IPAddress,
		UserAgent:   l.UserAgent,
		CreatedAt:   timestamppb.New(l.CreatedAt),
	}
}

func AdminMember(m adminuc.MemberOutput) *chatv1.AdminMember {
	member := &chatv1.AdminMember{
		UserId:             m.UserID,
		Email:              m.Email,
		DisplayName:        m.DisplayName,
		AvatarUrl:          m.AvatarURL,
		Role:               workspaceRoles[string(m.Role)],
		JoinedAt:           timestamppb.New(m.JoinedAt),
		SuspendedAt:        optionalTimestamp(m.SuspendedAt),
		RecentMessageCount: int32(m.Activity.MessageCount),
		StorageBytes:       m.Activity.StorageBytes,
		LastMessageAt:      optionalTimestamp(m.Activity.LastMessageAt),
	}
	if m.LastLogin != nil {
		member.LastLoginAt = timestamppb.New(m.LastLogin.CreatedAt)
		member.LastLoginIp = m.LastLogin.IPAddress
		member.LastLoginUserAgent = m.LastLogin.UserAgent
	}
	return member
}

func Permissions(out adminuc.PermissionsOutput) *chatv1.GetPermissionsResponse {
	res := &chatv1.GetPermissionsResponse{}
	for _, role := range entity.ConfigurableRoles {
		for _, p := range entity.AllPermissions {
			res.Grants = append(res.Grants, &chatv1.PermissionGrant{
				Role:       workspaceRoles[string(role)],
				Permission: permissions[p],
				Allowed:    out.Matrix.Allows(role, p),
			})
		}
	}
	for _, p := range entity.AllPermissions {
		if out.Matrix.Allows(out.RequesterRole, p) {
			res.MyPermissions = append(res.MyPermissions, permissions[p])
		}
	}
	return res
}
