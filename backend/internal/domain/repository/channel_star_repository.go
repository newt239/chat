package repository

import "context"

type ChannelStarRepository interface {
	SetStarred(ctx context.Context, userID string, channelID string, starred bool) error
	// FindStarredChannelIDs は channelIDs のうちスターを付けているものを返します
	FindStarredChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error)
}
