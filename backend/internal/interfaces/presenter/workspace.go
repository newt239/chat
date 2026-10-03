package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

var workspaceRoles = map[entity.WorkspaceRole]chatv1.WorkspaceRole{
	entity.WorkspaceRoleOwner:  chatv1.WorkspaceRole_WORKSPACE_ROLE_OWNER,
	entity.WorkspaceRoleAdmin:  chatv1.WorkspaceRole_WORKSPACE_ROLE_ADMIN,
	entity.WorkspaceRoleMember: chatv1.WorkspaceRole_WORKSPACE_ROLE_MEMBER,
	entity.WorkspaceRoleGuest:  chatv1.WorkspaceRole_WORKSPACE_ROLE_GUEST,
}

// WorkspaceRoleFromProto はリクエストのロールをエンティティのロールに変換します
func WorkspaceRoleFromProto(role chatv1.WorkspaceRole) entity.WorkspaceRole {
	return reverseLookup(workspaceRoles, role)
}

func Workspace(w workspaceuc.WorkspaceOutput) *chatv1.Workspace {
	return &chatv1.Workspace{
		Id:                 w.ID,
		Name:               w.Name,
		Description:        w.Description,
		IconUrl:            w.IconURL,
		IsPublic:           w.IsPublic,
		SignupEnabled:      w.SignupEnabled,
		EmailSignupEnabled: w.EmailSignupEnabled,
		Role:               workspaceRoles[w.Role],
		CreatedBy:          w.CreatedBy,
		CreatedAt:          timestamppb.New(w.CreatedAt),
		UpdatedAt:          timestamppb.New(w.UpdatedAt),
	}
}

func WorkspaceMember(m workspaceuc.MemberInfo) *chatv1.WorkspaceMember {
	return &chatv1.WorkspaceMember{
		UserId:      m.UserID,
		Email:       m.Email,
		DisplayName: m.DisplayName,
		AvatarUrl:   m.AvatarURL,
		Bio:         m.Bio,
		Role:        workspaceRoles[m.Role],
		JoinedAt:    timestamppb.New(m.JoinedAt),
		SuspendedAt: optionalTimestamp(m.SuspendedAt),
		Nickname:    m.Nickname,
		Timezone:    m.Timezone,
		Links:       m.Links,
	}
}

func PublicWorkspace(w workspaceuc.PublicWorkspaceItem) *chatv1.PublicWorkspace {
	return &chatv1.PublicWorkspace{
		Id:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		IconUrl:     w.IconURL,
		MemberCount: int32(w.MemberCount),
		IsJoined:    w.IsJoined,
		CreatedAt:   timestamppb.New(w.CreatedAt),
	}
}
