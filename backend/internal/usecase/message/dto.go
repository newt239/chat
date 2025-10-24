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

type UserMention struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
}

type GroupMention struct {
	GroupID string `json:"groupId"`
	Name    string `json:"name"`
}

type LinkInfo struct {
	ID          string  `json:"id"`
	URL         string  `json:"url"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ImageURL    *string `json:"imageUrl"`
	SiteName    *string `json:"siteName"`
	CardType    *string `json:"cardType"`
}

type ReactionInfo struct {
	User      UserInfo  `json:"user"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
}

type MessageOutput struct {
	ID        string         `json:"id"`
	ChannelID string         `json:"channelId"`
	UserID    string         `json:"userId"`
	User      UserInfo       `json:"user"`
	ParentID  *string        `json:"parentId"`
	Body      string         `json:"body"`
	Mentions  []UserMention  `json:"mentions"`
	Groups    []GroupMention `json:"groups"`
	Links     []LinkInfo     `json:"links"`
	Reactions []ReactionInfo `json:"reactions"`
	CreatedAt time.Time      `json:"createdAt"`
	EditedAt  *time.Time     `json:"editedAt"`
	DeletedAt *time.Time     `json:"deletedAt"`
}

type ListMessagesOutput struct {
	Messages []MessageOutput `json:"messages"`
	HasMore  bool            `json:"hasMore"`
}
