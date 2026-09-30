package entity

import (
	"regexp"
	"time"
)

var customEmojiNamePattern = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

// CustomEmoji はワークスペースで登録した絵文字です。本文やリアクションでは :name: と書きます
type CustomEmoji struct {
	ID          string
	WorkspaceID string
	Name        string
	StorageKey  string
	CreatedBy   string
	CreatedAt   time.Time
}

func IsValidCustomEmojiName(name string) bool {
	return customEmojiNamePattern.MatchString(name)
}
