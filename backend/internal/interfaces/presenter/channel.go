package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	channeluc "github.com/newt239/chat/internal/usecase/channel"
	channelmemberuc "github.com/newt239/chat/internal/usecase/channelmember"
)

// BrowsableChannelMembership はチャンネル一覧の参加状態の絞り込みを変換します
func BrowsableChannelMembership(m chatv1.BrowsableChannelMembership) domainrepository.BrowsableChannelMembership {
	switch m {
	case chatv1.BrowsableChannelMembership_BROWSABLE_CHANNEL_MEMBERSHIP_JOINED:
		return domainrepository.BrowsableChannelMembershipJoined
	case chatv1.BrowsableChannelMembership_BROWSABLE_CHANNEL_MEMBERSHIP_NOT_JOINED:
		return domainrepository.BrowsableChannelMembershipNotJoined
	default:
		return domainrepository.BrowsableChannelMembershipAll
	}
}

// BrowsableChannelSort はチャンネル一覧の並び順を変換します
func BrowsableChannelSort(s chatv1.BrowsableChannelSort) domainrepository.BrowsableChannelSort {
	if s == chatv1.BrowsableChannelSort_BROWSABLE_CHANNEL_SORT_MEMBER_COUNT {
		return domainrepository.BrowsableChannelSortMemberCount
	}
	return domainrepository.BrowsableChannelSortName
}

var channelRoles = map[entity.ChannelRole]chatv1.ChannelRole{
	entity.ChannelRoleMember: chatv1.ChannelRole_CHANNEL_ROLE_MEMBER,
	entity.ChannelRoleAdmin:  chatv1.ChannelRole_CHANNEL_ROLE_ADMIN,
}

// ChannelRoleFromProto はリクエストのロールをエンティティのロールに変換します
func ChannelRoleFromProto(role chatv1.ChannelRole) entity.ChannelRole {
	return reverseLookup(channelRoles, role)
}

func Channel(c channeluc.ChannelOutput) *chatv1.Channel {
	return &chatv1.Channel{
		Id:            c.ID,
		WorkspaceId:   c.WorkspaceID,
		Name:          c.Name,
		Description:   c.Description,
		IsPrivate:     c.IsPrivate,
		CreatedBy:     c.CreatedBy,
		CreatedAt:     timestamppb.New(c.CreatedAt),
		UpdatedAt:     timestamppb.New(c.UpdatedAt),
		UnreadCount:   int32(c.UnreadCount),
		HasMention:    c.MentionCount > 0,
		MentionCount:  int32(c.MentionCount),
		ArchivedAt:    optionalTimestamp(c.ArchivedAt),
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
