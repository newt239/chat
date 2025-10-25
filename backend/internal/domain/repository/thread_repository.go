package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type ThreadRepository interface {
	FindMetadataByMessageID(ctx context.Context, messageID string) (*entity.ThreadMetadata, error)
	FindMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.ThreadMetadata, error)
	CreateOrUpdateMetadata(ctx context.Context, metadata *entity.ThreadMetadata) error
	IncrementReplyCount(ctx context.Context, messageID string, replyUserID string) error
	DeleteMetadata(ctx context.Context, messageID string) error
}
