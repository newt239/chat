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
	ChannelID     string
	UserID        string
	Body          string
	ParentID      *string
	AttachmentIDs []string
}

type UpdateMessageInput struct {
	MessageID string
	ChannelID string
	EditorID  string
	Body      string
}

type DeleteMessageInput struct {
	MessageID string
	ChannelID string
	ExecutorID string
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

type AttachmentInfo struct {
	ID        string `json:"id"`
	FileName  string `json:"fileName"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

type MessageOutput struct {
	ID          string         `json:"id"`
	ChannelID   string         `json:"channelId"`
	UserID      string         `json:"userId"`
	User        UserInfo       `json:"user"`
	ParentID    *string        `json:"parentId"`
	Body        string         `json:"body"`
	Mentions    []UserMention  `json:"mentions"`
	Groups      []GroupMention `json:"groups"`
	Links       []LinkInfo     `json:"links"`
	Reactions   []ReactionInfo `json:"reactions"`
	Attachments []AttachmentInfo `json:"attachments"`
	CreatedAt   time.Time      `json:"createdAt"`
	EditedAt    *time.Time     `json:"editedAt"`
	DeletedAt   *time.Time     `json:"deletedAt"`
	IsDeleted   bool           `json:"isDeleted"`
	DeletedBy   *UserInfo      `json:"deletedBy,omitempty"`
}

type ListMessagesOutput struct {
	Messages []MessageOutput `json:"messages"`
	HasMore  bool            `json:"hasMore"`
}

type ThreadMetadataOutput struct {
	MessageID          string     `json:"messageId"`
	ReplyCount         int        `json:"replyCount"`
	LastReplyAt        *time.Time `json:"lastReplyAt"`
	LastReplyUser      *UserInfo  `json:"lastReplyUser"`
	ParticipantUserIDs []string   `json:"participantUserIds"`
}

type GetThreadRepliesInput struct {
	MessageID string
	UserID    string
	Limit     int
}

type GetThreadRepliesOutput struct {
	ParentMessage MessageOutput   `json:"parentMessage"`
	Replies       []MessageOutput `json:"replies"`
	HasMore       bool            `json:"hasMore"`
}

type GetThreadMetadataInput struct {
	MessageID string
	UserID    string
}

type MessageWithThreadOutput struct {
	MessageOutput
	ThreadMetadata *ThreadMetadataOutput `json:"threadMetadata,omitempty"`
}
