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
}

type ChannelOutput struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspaceId"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsPrivate   bool      `json:"isPrivate"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
