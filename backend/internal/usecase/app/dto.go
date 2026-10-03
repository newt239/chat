package app

import (
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

// PostInput は着信 Webhook に届いた投稿です
type PostInput struct {
	Text string
	// 省略したらアプリの既定のチャンネルに投稿する
	ChannelID *string
	ParentID  *string
}

// Output の OutgoingSecret は管理できる人にだけ設定する
type Output struct {
	*entity.App
	Creator   messageuc.UserInfo
	CanManage bool
}

type CreateOutput struct {
	App   Output
	Token string
}
