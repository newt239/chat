package presenter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/newt239/chat/internal/domain/entity"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	useruc "github.com/newt239/chat/internal/usecase/user"
	usernoteuc "github.com/newt239/chat/internal/usecase/usernote"
)

func AuthUser(u authuc.UserInfo) *chatv1.User {
	return &chatv1.User{Id: u.ID, Email: u.Email, DisplayName: u.DisplayName, AvatarUrl: u.AvatarURL}
}

func Me(me *useruc.MeOutput) *chatv1.User {
	return &chatv1.User{Id: me.ID, Email: me.Email, DisplayName: me.DisplayName, AvatarUrl: me.AvatarURL, Bio: me.Bio, Links: me.Links, Preferences: Preferences(me.Preferences)}
}

var sidebarStyles = map[entity.SidebarStyle]chatv1.SidebarStyle{
	entity.SidebarStyleTinted: chatv1.SidebarStyle_SIDEBAR_STYLE_TINTED,
	entity.SidebarStyleLight:  chatv1.SidebarStyle_SIDEBAR_STYLE_LIGHT,
}

var colorModes = map[entity.ColorMode]chatv1.ColorMode{
	entity.ColorModeLight:  chatv1.ColorMode_COLOR_MODE_LIGHT,
	entity.ColorModeDark:   chatv1.ColorMode_COLOR_MODE_DARK,
	entity.ColorModeSystem: chatv1.ColorMode_COLOR_MODE_SYSTEM,
}

var notificationLevels = map[entity.NotificationLevel]chatv1.NotificationLevel{
	entity.NotificationLevelAll:      chatv1.NotificationLevel_NOTIFICATION_LEVEL_ALL,
	entity.NotificationLevelMentions: chatv1.NotificationLevel_NOTIFICATION_LEVEL_MENTIONS,
	entity.NotificationLevelNone:     chatv1.NotificationLevel_NOTIFICATION_LEVEL_NONE,
}

var channelSortOrders = map[entity.ChannelSortOrder]chatv1.ChannelSortOrder{
	entity.ChannelSortOrderDefault:        chatv1.ChannelSortOrder_CHANNEL_SORT_ORDER_DEFAULT,
	entity.ChannelSortOrderRecentActivity: chatv1.ChannelSortOrder_CHANNEL_SORT_ORDER_RECENT_ACTIVITY,
}

func Preferences(p entity.UserPreferences) *chatv1.UserPreferences {
	return &chatv1.UserPreferences{
		Theme: &chatv1.ThemePreference{
			Hue:     int32(p.ThemeHue),
			Chroma:  p.ThemeChroma,
			Sidebar: sidebarStyles[p.ThemeSidebar],
		},
		ColorMode:          colorModes[p.ColorMode],
		Locale:             p.Locale,
		NotificationLevel:  notificationLevels[p.NotificationLevel],
		Timezone:           p.Timezone,
		TimezoneAutoUpdate: p.TimezoneAutoUpdate,
		ChannelSortOrder:   channelSortOrders[p.ChannelSortOrder],
		HideJoinMessages:   p.HideJoinMessages,
	}
}

// PreferencesFromProto は protovalidate で検証済みの入力をエンティティに変換します
func PreferencesFromProto(p *chatv1.UserPreferences) entity.UserPreferences {
	out := entity.UserPreferences{
		ThemeHue:           int(p.GetTheme().GetHue()),
		ThemeChroma:        p.GetTheme().GetChroma(),
		Locale:             p.GetLocale(),
		Timezone:           p.GetTimezone(),
		TimezoneAutoUpdate: p.GetTimezoneAutoUpdate(),
		HideJoinMessages:   p.GetHideJoinMessages(),
	}
	out.ThemeSidebar = reverseLookup(sidebarStyles, p.GetTheme().GetSidebar())
	out.ColorMode = reverseLookup(colorModes, p.GetColorMode())
	out.ChannelSortOrder = reverseLookup(channelSortOrders, p.GetChannelSortOrder())
	out.NotificationLevel = reverseLookup(notificationLevels, p.GetNotificationLevel())
	return out
}

// UserNote は未設定 (nil) の場合 nil を返します
func UserNote(n *usernoteuc.Output) *chatv1.UserNote {
	if n == nil {
		return nil
	}
	return &chatv1.UserNote{TargetUserId: n.TargetID, Nickname: n.Nickname, Memo: n.Memo, UpdatedAt: timestamppb.New(n.UpdatedAt)}
}

// ProfileLinksFromProto はリンクを指定しなかったときに nil を返し、変えないことを表します
func ProfileLinksFromProto(links *chatv1.ProfileLinks) *[]string {
	if links == nil {
		return nil
	}
	return &links.Urls
}
