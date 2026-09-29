package entity

import "time"

// Draft は入力途中のメッセージです。ParentID はスレッドへの返信のときだけ設定されます
type Draft struct {
	ID        string
	UserID    string
	ChannelID string
	ParentID  *string
	Body      string
	UpdatedAt time.Time
}
