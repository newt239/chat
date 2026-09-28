package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
)

var workspaceRoles = map[string]chatv1.WorkspaceRole{
	"owner":  chatv1.WorkspaceRole_WORKSPACE_ROLE_OWNER,
	"admin":  chatv1.WorkspaceRole_WORKSPACE_ROLE_ADMIN,
	"member": chatv1.WorkspaceRole_WORKSPACE_ROLE_MEMBER,
	"guest":  chatv1.WorkspaceRole_WORKSPACE_ROLE_GUEST,
}

// WorkspaceRoleName はリクエストのロールをユースケースが扱う文字列に変換します
func WorkspaceRoleName(role chatv1.WorkspaceRole) string {
	for name, r := range workspaceRoles {
		if r == role {
			return name
		}
	}
	return ""
}

func Workspace(w workspaceuc.WorkspaceOutput) *chatv1.Workspace {
	return &chatv1.Workspace{
		Id:          w.ID,
		Name:        w.Name,
		Description: w.Description,
		IconUrl:     w.IconURL,
		IsPublic:    w.IsPublic,
		Role:        workspaceRoles[w.Role],
		CreatedBy:   w.CreatedBy,
		CreatedAt:   timestamppb.New(w.CreatedAt),
		UpdatedAt:   timestamppb.New(w.UpdatedAt),
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
		Nickname:    m.Nickname,
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
