package entity

import "time"

type Message struct {
	ID              string
	ChannelID       string
	UserID          string
	ParentID        *string
	Body            string
	CreatedAt       time.Time
	EditedAt        *time.Time
	DeletedAt       *time.Time
	DeletedBy       *string
	Location        *MessageLocation
	MentionsChannel bool
	MentionsHere    bool
}

// CanBeRepliedIn は channelID のスレッドの親にできるメッセージかを返します。返信と削除済みのメッセージは親にできない
func (m *Message) CanBeRepliedIn(channelID string) bool {
	return m != nil && m.ChannelID == channelID && m.ParentID == nil && m.DeletedAt == nil
}

// MessageLocation はメッセージで共有された位置情報です
type MessageLocation struct {
	Latitude  float64
	Longitude float64
	// 測位の誤差（メートル）
	AccuracyMeters *float64
	Label          *string
}

type MessageReaction struct {
	MessageID string
	UserID    string
	Emoji     string
	CreatedAt time.Time
}
