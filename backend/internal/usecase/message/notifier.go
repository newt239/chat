package message

import (
	"context"
	"time"
)

// SearchIndexer はメッセージの変更を全文検索インデックスへ反映します。失敗しても呼び出し元へは返さない
type SearchIndexer interface {
	Sync(ctx context.Context, messageIDs ...string)
}

// Notifier はメッセージの変更をチャンネルの参加者へ配信します
type Notifier interface {
	NotifyNewMessage(workspaceID, channelID string, message MessageOutput)
	NotifyUpdatedMessage(workspaceID, channelID string, message MessageOutput)
	NotifyDeletedMessage(workspaceID, channelID string, deletion MessageDeletion)
}

// MessageDeletion は削除されたメッセージの通知内容です
type MessageDeletion struct {
	MessageID string
	// 親メッセージと一緒に削除されたスレッド返信も含む
	DeletedIDs []string
	DeletedAt  time.Time
}
