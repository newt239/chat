package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

var WorkspaceRoles = map[entity.WorkspaceRole]chatv1.WorkspaceRole{
	entity.WorkspaceRoleOwner:  chatv1.WorkspaceRole_WORKSPACE_ROLE_OWNER,
	entity.WorkspaceRoleAdmin:  chatv1.WorkspaceRole_WORKSPACE_ROLE_ADMIN,
	entity.WorkspaceRoleMember: chatv1.WorkspaceRole_WORKSPACE_ROLE_MEMBER,
	entity.WorkspaceRoleGuest:  chatv1.WorkspaceRole_WORKSPACE_ROLE_GUEST,
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
		Role:               WorkspaceRoles[w.Role],
		CreatedBy:          w.CreatedBy,
		CreatedAt:          timestamppb.New(w.CreatedAt),
		UpdatedAt:          timestamppb.New(w.UpdatedAt),
	}
}

func WorkspaceMember(m workspaceuc.MemberInfo) *chatv1.WorkspaceMember {
	return &chatv1.WorkspaceMember{
		UserId:      m.UserID,
		Email:       m.User.Email,
		DisplayName: m.User.DisplayName,
		AvatarUrl:   m.User.AvatarURL,
		Bio:         m.User.Bio,
		Role:        WorkspaceRoles[m.Role],
		SuspendedAt: optionalTimestamp(m.SuspendedAt),
		Nickname:    m.Nickname,
		Timezone:    m.User.Preferences.Timezone,
		Links:       m.User.Links,
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
