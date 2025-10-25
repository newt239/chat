package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ReadStateRepository interface {
	FindByChannelAndUser(ctx context.Context, channelID string, userID string) (*entity.ChannelReadState, error)
	Upsert(ctx context.Context, readState *entity.ChannelReadState) error
	GetUnreadCount(ctx context.Context, channelID string, userID string) (int, error)
	GetUnreadChannels(ctx context.Context, userID string) (map[string]int, error)
}
