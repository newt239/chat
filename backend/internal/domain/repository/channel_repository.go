package repository

import (
	"context"

	"github.com/example/chat/internal/domain/entity"
)

type ChannelRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Channel, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error)
	FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	Create(ctx context.Context, channel *entity.Channel) error
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *entity.ChannelMember) error
	RemoveMember(ctx context.Context, channelID string, userID string) error
	FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error)
	IsMember(ctx context.Context, channelID string, userID string) (bool, error)
}
