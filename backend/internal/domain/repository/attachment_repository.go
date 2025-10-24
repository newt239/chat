package repository

import (
	"context"

	"github.com/example/chat/internal/domain/entity"
)

type AttachmentRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Attachment, error)
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.Attachment, error)
	Create(ctx context.Context, attachment *entity.Attachment) error
	Delete(ctx context.Context, id string) error
}
