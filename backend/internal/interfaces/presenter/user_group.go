package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	usergroupuc "github.com/newt239/chat/internal/usecase/user_group"
)

func UserGroup(g usergroupuc.UserGroupOutput) *chatv1.UserGroup {
	return &chatv1.UserGroup{
		Id:          g.ID,
		WorkspaceId: g.WorkspaceID,
		Name:        g.Name,
		Description: g.Description,
		CreatedBy:   g.CreatedBy,
		CreatedAt:   timestamppb.New(g.CreatedAt),
		UpdatedAt:   timestamppb.New(g.UpdatedAt),
	}
}

func UserGroupMember(m usergroupuc.MemberInfo) *chatv1.UserGroupMember {
	return &chatv1.UserGroupMember{
		UserId:      m.UserID,
		DisplayName: m.DisplayName,
		AvatarUrl:   m.AvatarURL,
		JoinedAt:    timestamppb.New(m.JoinedAt),
	}
}
