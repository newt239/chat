package readstate

// Notifier は未読数の変化をユーザーへ配信します
type Notifier interface {
	NotifyUnreadCount(workspaceID, userID, channelID string, unreadCount int, hasMention bool)
}
