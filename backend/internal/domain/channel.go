package domain

import "time"

type Channel struct {
	ID          string
	WorkspaceID string
	Name        string
	Description *string
	IsPrivate   bool
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ChannelMember struct {
	ChannelID string
	UserID    string
	JoinedAt  time.Time
}

type ChannelRepository interface {
	FindByID(id string) (*Channel, error)
	FindByWorkspaceID(workspaceID string) ([]*Channel, error)
	FindAccessibleChannels(workspaceID, userID string) ([]*Channel, error)
	Create(channel *Channel) error
	Update(channel *Channel) error
	Delete(id string) error
	AddMember(member *ChannelMember) error
	RemoveMember(channelID, userID string) error
	FindMembers(channelID string) ([]*ChannelMember, error)
	IsMember(channelID, userID string) (bool, error)
}
