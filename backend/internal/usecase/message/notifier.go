package message

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// SearchIndexer はメッセージの変更を全文検索インデックスへ反映します。失敗しても呼び出し元へは返さない
type SearchIndexer interface {
	Sync(ctx context.Context, messageIDs ...string)
}

// NewMessageObserver は新着メッセージをプッシュ通知やアプリの送信 Webhook で外へ知らせます
type NewMessageObserver interface {
	NotifyNewMessage(ctx context.Context, channel *entity.Channel, message MessageOutput) error
}

// Notifier はメッセージの変更をチャンネルの参加者へ配信します
type Notifier interface {
	NotifyNewMessage(workspaceID, channelID string, message MessageOutput)
	NotifyUpdatedMessage(workspaceID, channelID string, message MessageOutput)
	NotifyDeletedMessage(workspaceID, channelID string, deletion MessageDeletion)
	NotifySystemMessageCreated(workspaceID, channelID string, message *entity.SystemMessage)
}

// MessageDeletion は削除されたメッセージの通知内容です
type MessageDeletion struct {
	MessageID string
	// 親メッセージと一緒に削除されたスレッド返信も含む
	DeletedIDs []string
}
