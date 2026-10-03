package reaction

import (
	"time"

	"github.com/newt239/chat/internal/usecase/message"
)

// Notifier はリアクションの変更をチャンネルの参加者へ配信します
type Notifier interface {
	NotifyReactionAdded(workspaceID, channelID string, reaction ReactionNotification)
	NotifyReactionRemoved(workspaceID, channelID string, reaction ReactionNotification)
}

// ReactionNotification はリアクション通知の内容です
type ReactionNotification struct {
	MessageID string
	UserID    string
	Emoji     string
	// 追加のときのみ設定される
	User      *message.UserInfo
	CreatedAt time.Time
}
