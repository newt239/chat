package app

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type ListInput struct {
	WorkspaceID string
	UserID      string
}

type ListByChannelInput struct {
	ChannelID string
	UserID    string
}

// SettingsInput は作成・更新で指定できる項目です
type SettingsInput struct {
	Name             string
	Description      *string
	AvatarURL        *string
	Permissions      []entity.AppPermission
	DefaultChannelID *string
	OutgoingURL      *string
}

type CreateInput struct {
	WorkspaceID string
	UserID      string
	Settings    SettingsInput
}

type UpdateInput struct {
	AppID    string
	UserID   string
	Settings SettingsInput
}

type TargetInput struct {
	AppID  string
	UserID string
}

type ChannelInput struct {
	AppID     string
	ChannelID string
	UserID    string
}

// PostInput は着信 Webhook に届いた投稿です。Username と AvatarURL はこの投稿の表示だけを上書きします
type PostInput struct {
	AppID     string
	Token     string
	Text      string
	Username  *string
	AvatarURL *string
	// 省略したらアプリの既定のチャンネルに投稿する
	ChannelID *string
	ParentID  *string
}

type Output struct {
	ID               string
	WorkspaceID      string
	Name             string
	Description      *string
	AvatarURL        *string
	Permissions      []entity.AppPermission
	DefaultChannelID *string
	OutgoingURL      *string
	// 管理できる人にだけ返す
	OutgoingSecret *string
	IsOfficial     bool
	BotUserID      string
	CreatedBy      messageuc.UserInfo
	CreatedAt      time.Time
	LastUsedAt     *time.Time
	CanManage      bool
}

type CreateOutput struct {
	App   Output
	Token string
}
