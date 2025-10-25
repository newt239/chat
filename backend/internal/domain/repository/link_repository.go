package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageLinkRepository interface {
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageLink, error)
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageLink, error)
	FindByURL(ctx context.Context, url string) (*entity.MessageLink, error)
	Create(ctx context.Context, link *entity.MessageLink) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}
