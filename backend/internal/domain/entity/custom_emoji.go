package entity

import "time"

// CustomEmoji はワークスペースで登録した絵文字です。本文やリアクションでは :name: と書きます
type CustomEmoji struct {
	ID          string
	WorkspaceID string
	Name        string
	StorageKey  string
	CreatedBy   string
	CreatedAt   time.Time
}
