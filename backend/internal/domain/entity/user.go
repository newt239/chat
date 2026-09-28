package entity

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("ユーザーが見つかりません")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	Bio          *string
	AvatarURL    *string
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

// UserPreferences は端末をまたいで共有する表示設定です
type UserPreferences struct {
	ThemeHue     int
	ThemeChroma  float64
	ThemeSidebar SidebarStyle
	ColorMode    ColorMode
	Locale       string
}
