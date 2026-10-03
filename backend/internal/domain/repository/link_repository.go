package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageLinkRepository interface {
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageLink, error)
	// FindPreviewsByURLs は保存済みのプレビューを URL ごとに返します
	FindPreviewsByURLs(ctx context.Context, urls []string) (map[string]*entity.LinkPreview, error)
	// UpsertPreview は URL ごとのプレビューを上書きで保存し、preview.ID を設定します
	UpsertPreview(ctx context.Context, preview *entity.LinkPreview) error
	CreateBulk(ctx context.Context, links []*entity.MessageLink) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}
