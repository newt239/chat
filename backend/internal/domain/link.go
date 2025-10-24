package domain

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

type MessageLinkRepository interface {
	FindByMessageID(messageID string) ([]*MessageLink, error)
	FindByMessageIDs(messageIDs []string) ([]*MessageLink, error)
	FindByURL(url string) (*MessageLink, error)
	Create(link *MessageLink) error
	DeleteByMessageID(messageID string) error
}
