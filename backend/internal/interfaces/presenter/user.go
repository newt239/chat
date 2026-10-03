package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

func AuthUser(u *entity.User) *chatv1.User {
	return &chatv1.User{Id: u.ID, Email: u.Email, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL}
}

func Me(u *entity.User) *chatv1.User {
	return &chatv1.User{Id: u.ID, Email: u.Email, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL, Bio: u.Bio, Links: u.Links, Preferences: Preferences(u.Preferences)}
}

var SidebarStyles = map[entity.SidebarStyle]chatv1.SidebarStyle{
	entity.SidebarStyleTinted: chatv1.SidebarStyle_SIDEBAR_STYLE_TINTED,
	entity.SidebarStyleLight:  chatv1.SidebarStyle_SIDEBAR_STYLE_LIGHT,
}

var ColorModes = map[entity.ColorMode]chatv1.ColorMode{
	entity.ColorModeLight:  chatv1.ColorMode_COLOR_MODE_LIGHT,
	entity.ColorModeDark:   chatv1.ColorMode_COLOR_MODE_DARK,
	entity.ColorModeSystem: chatv1.ColorMode_COLOR_MODE_SYSTEM,
}

var NotificationLevels = map[entity.NotificationLevel]chatv1.NotificationLevel{
	entity.NotificationLevelAll:      chatv1.NotificationLevel_NOTIFICATION_LEVEL_ALL,
	entity.NotificationLevelMentions: chatv1.NotificationLevel_NOTIFICATION_LEVEL_MENTIONS,
	entity.NotificationLevelNone:     chatv1.NotificationLevel_NOTIFICATION_LEVEL_NONE,
}

var ChannelSortOrders = map[entity.ChannelSortOrder]chatv1.ChannelSortOrder{
	entity.ChannelSortOrderDefault:        chatv1.ChannelSortOrder_CHANNEL_SORT_ORDER_DEFAULT,
	entity.ChannelSortOrderRecentActivity: chatv1.ChannelSortOrder_CHANNEL_SORT_ORDER_RECENT_ACTIVITY,
}

func Preferences(p entity.UserPreferences) *chatv1.UserPreferences {
	return &chatv1.UserPreferences{
		Theme: &chatv1.ThemePreference{
			Hue:     int32(p.ThemeHue),
			Chroma:  p.ThemeChroma,
			Sidebar: SidebarStyles[p.ThemeSidebar],
		},
		ColorMode:          ColorModes[p.ColorMode],
		Locale:             p.Locale,
		NotificationLevel:  NotificationLevels[p.NotificationLevel],
		Timezone:           p.Timezone,
		TimezoneAutoUpdate: p.TimezoneAutoUpdate,
		ChannelSortOrder:   ChannelSortOrders[p.ChannelSortOrder],
		HideJoinMessages:   p.HideJoinMessages,
	}
}

// UserNote は未設定 (nil) の場合 nil を返します
func UserNote(n *entity.UserNote) *chatv1.UserNote {
	if n == nil {
		return nil
	}
	return &chatv1.UserNote{TargetUserId: n.TargetID, Nickname: n.Nickname, Memo: n.Memo, UpdatedAt: timestamppb.New(n.UpdatedAt)}
}

func UserGroup(g *entity.UserGroup) *chatv1.UserGroup {
	return &chatv1.UserGroup{
		Id:          g.ID,
		WorkspaceId: g.WorkspaceID,
		Name:        g.Name,
		Description: g.Description,
	}
}

func UserGroupMember(u *entity.User) *chatv1.UserGroupMember {
	return &chatv1.UserGroupMember{UserId: u.ID, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL}
}
