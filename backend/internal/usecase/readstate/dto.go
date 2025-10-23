package readstate

import "time"

type GetUnreadCountInput struct {
	ChannelID string
	UserID    string
}

type UpdateReadStateInput struct {
	ChannelID  string
	UserID     string
	LastReadAt time.Time
}

type UnreadCountOutput struct {
	Count int `json:"count"`
}
