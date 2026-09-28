package pin

import (
	"time"

	"github.com/newt239/chat/internal/usecase/message"
)

// Notifier はピン留めの変更をチャンネルの参加者へ配信します
type Notifier interface {
	NotifyPinCreated(workspaceID, channelID string, pin PinNotification)
	NotifyPinDeleted(workspaceID, channelID string, pin PinNotification)
}

// PinNotification はピン留め通知の内容です
type PinNotification struct {
	MessageID string
	PinnedBy  string
	PinnedAt  time.Time
	// ピン留めしたときのみ設定される
	PinnedByUser *message.UserInfo
}
