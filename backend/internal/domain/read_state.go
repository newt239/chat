package domain

import "time"

type ChannelReadState struct {
	ChannelID  string
	UserID     string
	LastReadAt time.Time
}

type ReadStateRepository interface {
	FindByChannelAndUser(channelID, userID string) (*ChannelReadState, error)
	Upsert(readState *ChannelReadState) error
	GetUnreadCount(channelID, userID string) (int, error)
	GetUnreadChannels(userID string) (map[string]int, error)
}
