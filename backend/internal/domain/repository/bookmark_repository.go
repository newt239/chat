package repository

import (
	"context"

	"github.com/example/chat/internal/domain/entity"
)

type BookmarkRepository interface {
	AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error
	RemoveBookmark(ctx context.Context, userID, messageID string) error
	FindByUserID(ctx context.Context, userID string) ([]*entity.MessageBookmark, error)
	IsBookmarked(ctx context.Context, userID, messageID string) (bool, error)
}
