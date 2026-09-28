package reaction

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
}
