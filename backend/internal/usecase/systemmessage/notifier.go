package systemmessage

import "github.com/newt239/chat/internal/domain/entity"

// Notifier はシステムメッセージの作成をチャンネルの参加者へ配信します
type Notifier interface {
	NotifySystemMessageCreated(workspaceID, channelID string, message *entity.SystemMessage)
}
