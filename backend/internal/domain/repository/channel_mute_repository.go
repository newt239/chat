package repository

import "context"

type ChannelMuteRepository interface {
	SetMuted(ctx context.Context, userID string, channelID string, muted bool) error
	// FindMutedChannelIDs は channelIDs のうちミュートしているものを返します
	FindMutedChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error)
}
