package channel

import "time"

type ListChannelsInput struct {
	WorkspaceID string
	UserID      string
}

type CreateChannelInput struct {
	WorkspaceID string
	UserID      string
	Name        string
	Description *string
	IsPrivate   bool
	MemberIDs   []string
}

type UpdateChannelInput struct {
	ChannelID   string
	UserID      string
	Name        *string
	Description *string
	IsPrivate   *bool
}

type GetChannelInput struct {
	ChannelID string
	UserID    string
}

type DeleteChannelInput struct {
	ChannelID string
	UserID    string
}

type SetArchivedInput struct {
	ChannelID string
	UserID    string
	Archived  bool
}

type SetChannelStarredInput struct {
	ChannelID string
	UserID    string
	Starred   bool
}

type ChannelOutput struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspaceId"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	IsPrivate   bool       `json:"isPrivate"`
	CreatedBy   string     `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	UnreadCount int        `json:"unreadCount"`
	HasMention  bool       `json:"hasMention"`
	ParentID    *string    `json:"parentId"`
	IsStarred   bool       `json:"isStarred"`
	IsMember    bool       `json:"isMember"`
	ArchivedAt  *time.Time `json:"archivedAt,omitempty"`
}
