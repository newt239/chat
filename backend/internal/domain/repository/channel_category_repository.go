package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelCategoryRepository interface {
	FindByID(ctx context.Context, id string) (*entity.ChannelCategory, error)
	// FindByUser は position の昇順で割り当てたチャンネルと一緒に返します
	FindByUser(ctx context.Context, userID string, workspaceID string) ([]*entity.ChannelCategory, error)
	Create(ctx context.Context, category *entity.ChannelCategory) error
	UpdateName(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
	// UpdatePositions は categoryIDs の並び順を position として保存します
	UpdatePositions(ctx context.Context, categoryIDs []string) error
	// SetChannel はチャンネルの割り当てを置き換えます。categoryID が nil なら割り当てを外します
	SetChannel(ctx context.Context, userID string, channelID string, categoryID *string) error
}
