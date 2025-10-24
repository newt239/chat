package entity

import "time"

type ChannelReadState struct {
	ChannelID  string
	UserID     string
	LastReadAt time.Time
}
