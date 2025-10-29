package service

// NotificationService はリアルタイム通知を管理するサービスです
type NotificationService interface {
	// NotifyNewMessage は新しいメッセージをチャンネル参加者に通知します
	NotifyNewMessage(workspaceID string, channelID string, message interface{})

	// NotifyUpdatedMessage はメッセージ更新をチャンネル参加者に通知します
	NotifyUpdatedMessage(workspaceID string, channelID string, message interface{})

	// NotifyDeletedMessage はメッセージ削除をチャンネル参加者に通知します
	NotifyDeletedMessage(workspaceID string, channelID string, deleteData interface{})

	// NotifyReaction はリアクション追加をチャンネル参加者に通知します
	NotifyReaction(workspaceID string, channelID string, reaction interface{})

	// NotifyUnreadCount は未読数の更新を特定ユーザーに通知します
	NotifyUnreadCount(workspaceID string, userID string, channelID string, unreadCount int)

	// ピン関連
	NotifyPinCreated(workspaceID string, channelID string, pin interface{})
	NotifyPinDeleted(workspaceID string, channelID string, pin interface{})
}
