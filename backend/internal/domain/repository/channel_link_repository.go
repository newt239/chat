package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelLinkRepository interface {
	FindByID(ctx context.Context, id string) (*entity.ChannelLink, error)
	// FindByChannelID は position の昇順でリンクを返します
	FindByChannelID(ctx context.Context, channelID string) ([]*entity.ChannelLink, error)
	Create(ctx context.Context, link *entity.ChannelLink) error
	Update(ctx context.Context, link *entity.ChannelLink) error
	Delete(ctx context.Context, id string) error
	// UpdatePositions は linkIDs の並び順を position として保存します
	UpdatePositions(ctx context.Context, linkIDs []string) error
}
