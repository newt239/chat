package entity

import "time"

// MessagePin はチャンネル内のメッセージのピン留めを表します
type MessagePin struct {
	ID        string
	ChannelID string
	MessageID string
	PinnedBy  string
	PinnedAt  time.Time

	// 一覧取得時にメッセージ要約を組み立てるために使用
	Message *Message
}
