package customemoji

// Notifier はカスタム絵文字の登録・削除をワークスペースの全員へ配信します
type Notifier interface {
	NotifyCustomEmojisChanged(workspaceID string)
}
