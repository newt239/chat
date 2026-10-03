package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	adminuc "github.com/newt239/chat/internal/usecase/admin"
)

var AuditActions = map[entity.AuditAction]chatv1.AuditAction{
	entity.AuditActionLogin:              chatv1.AuditAction_AUDIT_ACTION_LOGIN,
	entity.AuditActionLoginFailed:        chatv1.AuditAction_AUDIT_ACTION_LOGIN_FAILED,
	entity.AuditActionMemberRoleChanged:  chatv1.AuditAction_AUDIT_ACTION_MEMBER_ROLE_CHANGED,
	entity.AuditActionMemberSuspended:    chatv1.AuditAction_AUDIT_ACTION_MEMBER_SUSPENDED,
	entity.AuditActionMemberResumed:      chatv1.AuditAction_AUDIT_ACTION_MEMBER_RESUMED,
	entity.AuditActionChannelCreated:     chatv1.AuditAction_AUDIT_ACTION_CHANNEL_CREATED,
	entity.AuditActionPermissionChanged:  chatv1.AuditAction_AUDIT_ACTION_PERMISSION_CHANGED,
	entity.AuditActionAuditLogExported:   chatv1.AuditAction_AUDIT_ACTION_AUDIT_LOG_EXPORTED,
	entity.AuditActionCustomEmojiCreated: chatv1.AuditAction_AUDIT_ACTION_CUSTOM_EMOJI_CREATED,
	entity.AuditActionCustomEmojiDeleted: chatv1.AuditAction_AUDIT_ACTION_CUSTOM_EMOJI_DELETED,
	entity.AuditActionAppCreated:         chatv1.AuditAction_AUDIT_ACTION_APP_CREATED,
	entity.AuditActionAppDeleted:         chatv1.AuditAction_AUDIT_ACTION_APP_DELETED,
}

var Permissions = map[entity.Permission]chatv1.Permission{
	entity.PermissionCreatePublicChannel:  chatv1.Permission_PERMISSION_CREATE_PUBLIC_CHANNEL,
	entity.PermissionCreatePrivateChannel: chatv1.Permission_PERMISSION_CREATE_PRIVATE_CHANNEL,
	entity.PermissionInviteMembers:        chatv1.Permission_PERMISSION_INVITE_MEMBERS,
	entity.PermissionEditChannelLinks:     chatv1.Permission_PERMISSION_EDIT_CHANNEL_LINKS,
	entity.PermissionPinMessages:          chatv1.Permission_PERMISSION_PIN_MESSAGES,
	entity.PermissionDeleteOthersMessages: chatv1.Permission_PERMISSION_DELETE_OTHERS_MESSAGES,
	entity.PermissionCreateCustomEmoji:    chatv1.Permission_PERMISSION_CREATE_CUSTOM_EMOJI,
}

func AuditLog(l adminuc.AuditLogOutput) *chatv1.AuditLog {
	log := &chatv1.AuditLog{
		Id:          l.ID,
		Action:      AuditActions[l.Action],
		TargetType:  string(l.TargetType),
		TargetId:    l.TargetID,
		TargetLabel: l.TargetLabel,
		Metadata:    l.Metadata,
		IpAddress:   l.IPAddress,
		UserAgent:   l.UserAgent,
		CreatedAt:   timestamppb.New(l.CreatedAt),
	}
	if l.Actor != nil {
		log.Actor = UserSummary(*l.Actor)
	}
	return log
}

func AdminMember(m adminuc.MemberOutput) *chatv1.AdminMember {
	member := &chatv1.AdminMember{
		UserId:             m.UserID,
		Email:              m.Email,
		DisplayName:        m.DisplayName,
		AvatarUrl:          m.AvatarURL,
		Role:               WorkspaceRoles[m.Role],
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

func PermissionsResponse(out adminuc.PermissionsOutput) *chatv1.GetPermissionsResponse {
	res := &chatv1.GetPermissionsResponse{}
	for _, role := range entity.ConfigurableRoles {
		for _, p := range entity.AllPermissions {
			res.Grants = append(res.Grants, &chatv1.PermissionGrant{
				Role:       WorkspaceRoles[role],
				Permission: Permissions[p],
				Allowed:    out.Matrix.Allows(role, p),
			})
		}
	}
	for _, p := range entity.AllPermissions {
		if out.Matrix.Allows(out.RequesterRole, p) {
			res.MyPermissions = append(res.MyPermissions, Permissions[p])
		}
	}
	return res
}
