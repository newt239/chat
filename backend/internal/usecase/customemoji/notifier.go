package customemoji

// Notifier はカスタム絵文字の登録・削除をワークスペースの全員へ配信します
type Notifier interface {
	NotifyCustomEmojiCreated(workspaceID string, emoji Notification)
	NotifyCustomEmojiDeleted(workspaceID string, emoji Notification)
}

type Notification struct {
	ID   string
	Name string
}
