package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type CustomEmojiRepository interface {
	FindByID(ctx context.Context, id string) (*entity.CustomEmoji, error)
	// FindByWorkspaceID は名前の昇順で返します
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.CustomEmoji, error)
	// Create は同じワークスペースに同じ名前があると errors.ErrCustomEmojiNameExists を返します
	Create(ctx context.Context, emoji *entity.CustomEmoji) error
	Delete(ctx context.Context, id string) error
}
