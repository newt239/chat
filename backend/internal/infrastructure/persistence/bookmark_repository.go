package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/database"
)

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) domainrepository.BookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return resolveDB(ctx, r.db)
}

func (r *bookmarkRepository) AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error {
	model := &database.MessageBookmark{}
	model.FromEntity(bookmark)

	if err := r.dbWithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	bookmark.CreatedAt = model.CreatedAt
	return nil
}

func (r *bookmarkRepository) RemoveBookmark(ctx context.Context, userID, messageID string) error {
	return r.dbWithContext(ctx).
		Delete(&database.MessageBookmark{}, "user_id = ? AND message_id = ?", userID, messageID).
		Error
}

func (r *bookmarkRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.MessageBookmark, error) {
	var models []database.MessageBookmark
	if err := r.dbWithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	bookmarks := make([]*entity.MessageBookmark, len(models))
	for i := range models {
		bookmarks[i] = models[i].ToEntity()
	}

	return bookmarks, nil
}

func (r *bookmarkRepository) IsBookmarked(ctx context.Context, userID, messageID string) (bool, error) {
	var count int64
	if err := r.dbWithContext(ctx).
		Model(&database.MessageBookmark{}).
		Where("user_id = ? AND message_id = ?", userID, messageID).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
