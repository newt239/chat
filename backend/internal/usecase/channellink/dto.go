package channellink

import "time"

type ListInput struct {
	ChannelID string
	UserID    string
}

type CreateInput struct {
	ChannelID string
	UserID    string
	Title     string
	URL       string
}

type UpdateInput struct {
	LinkID string
	UserID string
	Title  string
	URL    string
}

type DeleteInput struct {
	LinkID string
	UserID string
}

type ReorderInput struct {
	ChannelID string
	UserID    string
	LinkIDs   []string
}

type LinkOutput struct {
	ID        string
	ChannelID string
	Title     string
	URL       string
	Position  int
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListOutput struct {
	Links   []LinkOutput
	CanEdit bool
}
