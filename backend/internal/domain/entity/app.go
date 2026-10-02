package entity

import (
	"crypto/subtle"
	"slices"
	"time"
)

// AppPermission はアプリに許す操作です
type AppPermission string

const (
	// AppPermissionPostJoinedChannels は参加しているチャンネルに投稿できる
	AppPermissionPostJoinedChannels AppPermission = "post:joined_channels"
	// AppPermissionPostPublicChannels は参加していなくても公開チャンネルに投稿できる
	AppPermissionPostPublicChannels AppPermission = "post:public_channels"
	// AppPermissionPostThreadReplies はスレッドに返信できる
	AppPermissionPostThreadReplies AppPermission = "post:thread_replies"
	// AppPermissionOutgoingWebhook は参加しているチャンネルに投稿があると外部へ送る
	AppPermissionOutgoingWebhook AppPermission = "webhook:outgoing"
)

var AllAppPermissions = []AppPermission{
	AppPermissionPostJoinedChannels,
	AppPermissionPostPublicChannels,
	AppPermissionPostThreadReplies,
	AppPermissionOutgoingWebhook,
}

// App はワークスペースの連携アプリです。投稿は BotUserID の名義で行う
type App struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	AvatarURL   *string
	// 着信 Webhook のトークンのハッシュ。トークンは発行時にだけ平文で返す
	TokenHash   *string
	Permissions []AppPermission
	// 着信 Webhook で投稿先を省略したときのチャンネル
	DefaultChannelID *string
	OutgoingURL      *string
	// 送信 Webhook の署名に使う。受け取り側が検証できるよう管理者には見せる
	OutgoingSecret *string
	IsOfficial     bool
	BotUserID      string
	CreatedBy      string
	LastUsedAt     *time.Time
	CreatedAt      time.Time
}

func (a *App) Has(permission AppPermission) bool {
	return slices.Contains(a.Permissions, permission)
}

func (a *App) VerifyToken(token string) bool {
	return a.TokenHash != nil && subtle.ConstantTimeCompare([]byte(HashSecretToken(token)), []byte(*a.TokenHash)) == 1
}

// CanPostTo はアプリがチャンネルに投稿できるかを返します。isMember はボットユーザーがチャンネルに参加しているか
func (a *App) CanPostTo(channel *Channel, isMember bool) bool {
	if a.IsOfficial {
		return true
	}
	if isMember && a.Has(AppPermissionPostJoinedChannels) {
		return true
	}
	return channel.Type == ChannelTypePublic && a.Has(AppPermissionPostPublicChannels)
}
