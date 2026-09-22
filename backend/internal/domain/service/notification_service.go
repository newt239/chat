package service

// NotificationService はリアルタイム通知を管理するサービスです
type NotificationService interface {
	// NotifyNewMessage は新しいメッセージをチャンネル参加者に通知します
	NotifyNewMessage(workspaceID string, channelID string, message interface{})

	// NotifyUpdatedMessage はメッセージ更新をチャンネル参加者に通知します
	NotifyUpdatedMessage(workspaceID string, channelID string, message interface{})

	// NotifyDeletedMessage はメッセージ削除をチャンネル参加者に通知します
	NotifyDeletedMessage(workspaceID string, channelID string, deleteData interface{})

	// NotifyReactionAdded はリアクション追加をチャンネル参加者に通知します
	NotifyReactionAdded(workspaceID string, channelID string, reaction ReactionNotification)

	// NotifyReactionRemoved はリアクション削除をチャンネル参加者に通知します
	NotifyReactionRemoved(workspaceID string, channelID string, reaction ReactionNotification)

	// NotifyUnreadCount は未読数の更新を特定ユーザーに通知します
	NotifyUnreadCount(workspaceID string, userID string, channelID string, unreadCount int, hasMention bool)

	// ピン関連
	NotifyPinCreated(workspaceID string, channelID string, pin interface{})
	NotifyPinDeleted(workspaceID string, channelID string, pin interface{})

	// システムメッセージ関連
	NotifySystemMessageCreated(workspaceID string, channelID string, message interface{})
}

// ReactionNotification はリアクション通知の内容です
type ReactionNotification struct {
	MessageID string
	UserID    string
	Emoji     string
}
