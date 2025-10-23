package message

import "time"

type ListMessagesInput struct {
	ChannelID string
	UserID    string
	Limit     int
	Since     *time.Time
	Until     *time.Time
}

type CreateMessageInput struct {
	ChannelID string
	UserID    string
	Body      string
	ParentID  *string
}

type UserInfo struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type MessageOutput struct {
	ID        string     `json:"id"`
	ChannelID string     `json:"channelId"`
	UserID    string     `json:"userId"`
	User      UserInfo   `json:"user"`
	ParentID  *string    `json:"parentId"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
	EditedAt  *time.Time `json:"editedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
}

type ListMessagesOutput struct {
	Messages []MessageOutput `json:"messages"`
	HasMore  bool            `json:"hasMore"`
}
