package entity

import "time"

type ChannelLink struct {
	ID        string
	ChannelID string
	Title     string
	URL       string
	Position  int
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}
