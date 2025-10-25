package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Channel, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error)
	FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	Create(ctx context.Context, channel *entity.Channel) error
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id string) error
}
