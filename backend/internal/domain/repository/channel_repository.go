package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Channel, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error)
	FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	SearchAccessibleChannels(ctx context.Context, workspaceID string, userID string, query string, limit int, offset int) ([]*entity.Channel, int, error)
	Create(ctx context.Context, channel *entity.Channel) error
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id string) error
	FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error)
	FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string, name string) (*entity.Channel, error)
	FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
}
