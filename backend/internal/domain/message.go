package domain

import "time"

type Message struct {
	ID        string
	ChannelID string
	UserID    string
	ParentID  *string // For thread support
	Body      string
	CreatedAt time.Time
	EditedAt  *time.Time
	DeletedAt *time.Time
}

type MessageReaction struct {
	MessageID string
	UserID    string
	Emoji     string
	CreatedAt time.Time
}

type MessageRepository interface {
	FindByID(id string) (*Message, error)
	FindByChannelID(channelID string, limit int, since, until *time.Time) ([]*Message, error)
	FindThreadReplies(parentID string) ([]*Message, error)
	Create(message *Message) error
	Update(message *Message) error
	Delete(id string) error
	AddReaction(reaction *MessageReaction) error
	RemoveReaction(messageID, userID, emoji string) error
	FindReactions(messageID string) ([]*MessageReaction, error)
}
