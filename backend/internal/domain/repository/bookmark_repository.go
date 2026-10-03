package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type BookmarkRepository interface {
	// AddBookmark は既にブックマークしていれば ErrBookmarkExists を返します
	AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error
	RemoveBookmark(ctx context.Context, userID, messageID string) error
	// FindByUserID はワークスペース内の削除されていないメッセージのブックマークを新しい順に返します
	FindByUserID(ctx context.Context, userID, workspaceID string) ([]*entity.MessageBookmark, error)
}
