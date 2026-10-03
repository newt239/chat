package customemoji

import "github.com/newt239/chat/internal/domain/entity"

// Notifier はカスタム絵文字の登録・削除をワークスペースの全員へ配信します
type Notifier interface {
	NotifyCustomEmojiCreated(emoji *entity.CustomEmoji)
	NotifyCustomEmojiDeleted(emoji *entity.CustomEmoji)
}
