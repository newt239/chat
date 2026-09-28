package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
)

var channelRoles = map[string]chatv1.ChannelRole{
	"member": chatv1.ChannelRole_CHANNEL_ROLE_MEMBER,
	"admin":  chatv1.ChannelRole_CHANNEL_ROLE_ADMIN,
}

// ChannelRoleName はリクエストのロールをユースケースが扱う文字列に変換します
func ChannelRoleName(role chatv1.ChannelRole) string {
	for name, r := range channelRoles {
		if r == role {
			return name
		}
	}
	return ""
}

func Channel(c channeluc.ChannelOutput) *chatv1.Channel {
	return &chatv1.Channel{
		Id:          c.ID,
		WorkspaceId: c.WorkspaceID,
		Name:        c.Name,
		Description: c.Description,
		IsPrivate:   c.IsPrivate,
		CreatedBy:   c.CreatedBy,
		CreatedAt:   timestamppb.New(c.CreatedAt),
		UpdatedAt:   timestamppb.New(c.UpdatedAt),
		UnreadCount: int32(c.UnreadCount),
		HasMention:  c.HasMention,
		ParentId:    c.ParentID,
		IsStarred:   c.IsStarred,
		IsMember:    c.IsMember,
	}
}

func ChannelMember(m channelmemberuc.MemberInfo) *chatv1.ChannelMember {
	return &chatv1.ChannelMember{
		UserId:      m.UserID,
		Email:       m.Email,
		DisplayName: m.DisplayName,
		AvatarUrl:   m.AvatarURL,
		Role:        channelRoles[m.Role],
		JoinedAt:    timestamppb.New(m.JoinedAt),
	}
}
