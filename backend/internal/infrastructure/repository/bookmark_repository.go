package repository

import (
	"context"
	"fmt"

	"github.com/example/chat/internal/domain/entity"
	"gorm.io/gorm"
)

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) *bookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error {
	if err := r.db.WithContext(ctx).Create(bookmark).Error; err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	return nil
}

func (r *bookmarkRepository) RemoveBookmark(ctx context.Context, userID, messageID string) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND message_id = ?", userID, messageID).Delete(&entity.MessageBookmark{}).Error; err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}
	return nil
}

func (r *bookmarkRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.MessageBookmark, error) {
	var bookmarks []*entity.MessageBookmark
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&bookmarks).Error; err != nil {
		return nil, fmt.Errorf("failed to find bookmarks: %w", err)
	}
	return bookmarks, nil
}

func (r *bookmarkRepository) IsBookmarked(ctx context.Context, userID, messageID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.MessageBookmark{}).Where("user_id = ? AND message_id = ?", userID, messageID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check bookmark status: %w", err)
	}
	return count > 0, nil
}
