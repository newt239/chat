package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

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
