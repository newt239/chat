package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
)

var ChannelRoles = map[entity.ChannelRole]chatv1.ChannelRole{
	entity.ChannelRoleMember: chatv1.ChannelRole_CHANNEL_ROLE_MEMBER,
	entity.ChannelRoleAdmin:  chatv1.ChannelRole_CHANNEL_ROLE_ADMIN,
}

func Channel(c channeluc.ChannelOutput) *chatv1.Channel {
	return &chatv1.Channel{
		Id:            c.ID,
		WorkspaceId:   c.WorkspaceID,
		Name:          c.Name,
		Description:   c.Description,
		IsPrivate:     c.IsPrivate(),
		CreatedBy:     c.CreatedBy,
		CreatedAt:     timestamppb.New(c.CreatedAt),
		UpdatedAt:     timestamppb.New(c.UpdatedAt),
		UnreadCount:   int32(c.UnreadCount),
		MentionCount:  int32(c.MentionCount),
		ParentId:      c.ParentID,
		IsStarred:     c.IsStarred,
		IsMuted:       c.IsMuted,
		IsMember:      c.IsMember,
		LastMessageAt: optionalTimestamp(c.LastMessageAt),
	}
}

func BrowsableChannel(c channeluc.BrowsableChannelOutput) *chatv1.BrowsableChannel {
	return &chatv1.BrowsableChannel{Channel: Channel(c.Channel), MemberCount: int32(c.MemberCount)}
}

func ChannelMember(m channelmemberuc.MemberOutput) *chatv1.ChannelMember {
	return &chatv1.ChannelMember{
		UserId:      m.User.ID,
		Email:       m.User.Email,
		DisplayName: m.User.DisplayName,
		AvatarUrl:   m.User.AvatarURL,
		Role:        ChannelRoles[m.Role],
	}
}

func ChannelLink(l *entity.ChannelLink) *chatv1.ChannelLink {
	return &chatv1.ChannelLink{
		Id:        l.ID,
		ChannelId: l.ChannelID,
		Title:     l.Title,
		Url:       l.URL,
		Position:  int32(l.Position),
		CreatedBy: l.CreatedBy,
		CreatedAt: timestamppb.New(l.CreatedAt),
		UpdatedAt: timestamppb.New(l.UpdatedAt),
	}
}

func ChannelCategory(c *entity.ChannelCategory) *chatv1.ChannelCategory {
	return &chatv1.ChannelCategory{Id: c.ID, Name: c.Name, Position: int32(c.Position), ChannelIds: c.ChannelIDs}
}
