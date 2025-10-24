package domain

import "time"

type MessageUserMention struct {
	MessageID string
	UserID    string
	CreatedAt time.Time
}

type MessageGroupMention struct {
	MessageID string
	GroupID   string
	CreatedAt time.Time
}

type MessageUserMentionRepository interface {
	FindByMessageID(messageID string) ([]*MessageUserMention, error)
	FindByMessageIDs(messageIDs []string) ([]*MessageUserMention, error)
	FindByUserID(userID string, limit int, since *time.Time) ([]*MessageUserMention, error)
	Create(mention *MessageUserMention) error
	DeleteByMessageID(messageID string) error
}

type MessageGroupMentionRepository interface {
	FindByMessageID(messageID string) ([]*MessageGroupMention, error)
	FindByMessageIDs(messageIDs []string) ([]*MessageGroupMention, error)
	FindByGroupID(groupID string, limit int, since *time.Time) ([]*MessageGroupMention, error)
	Create(mention *MessageGroupMention) error
	DeleteByMessageID(messageID string) error
}
