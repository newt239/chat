package entity

import "time"

type MessageLink struct {
	ID          string
	MessageID   string
	URL         string
	Title       *string
	Description *string
	ImageURL    *string
	SiteName    *string
	CardType    *string
	CreatedAt   time.Time
}
