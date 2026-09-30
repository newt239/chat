package entity

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("ユーザーが見つかりません")
)

// UnusablePasswordHash はパスワードでログインできないユーザー（Google アカウントのみ・ボット）に設定します
const UnusablePasswordHash = "!"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	GoogleSub    *string
	DisplayName  string
	Bio          *string
	AvatarURL    *string
	IsBot        bool
	Preferences  UserPreferences
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type SidebarStyle string

const (
	SidebarStyleTinted SidebarStyle = "tinted"
	SidebarStyleLight  SidebarStyle = "light"
)

type ColorMode string

const (
	ColorModeLight  ColorMode = "light"
	ColorModeDark   ColorMode = "dark"
	ColorModeSystem ColorMode = "system"
)

// NotificationLevel は通知を受け取る範囲です
type NotificationLevel string

const (
	NotificationLevelAll      NotificationLevel = "all"
	NotificationLevelMentions NotificationLevel = "mentions"
	NotificationLevelNone     NotificationLevel = "none"
)

type ChannelSortOrder string

const (
	ChannelSortOrderDefault        ChannelSortOrder = "default"
	ChannelSortOrderRecentActivity ChannelSortOrder = "recent_activity"
)

// UserPreferences は端末をまたいで共有する表示・通知の設定です
type UserPreferences struct {
	ThemeHue          int
	ThemeChroma       float64
	ThemeSidebar      SidebarStyle
	ColorMode         ColorMode
	Locale            string
	NotificationLevel NotificationLevel
	// IANA のタイムゾーン名。空は未設定
	Timezone           string
	TimezoneAutoUpdate bool
	ChannelSortOrder   ChannelSortOrder
	HideJoinMessages   bool
}
