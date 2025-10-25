package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type AttachmentRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Attachment, error)
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.Attachment, error)
	FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.Attachment, error)
	FindPendingByIDsForUser(ctx context.Context, userID string, attachmentIDs []string) ([]*entity.Attachment, error)
	CreatePending(ctx context.Context, attachment *entity.Attachment) error
	AttachToMessage(ctx context.Context, attachmentIDs []string, messageID string) error
	Delete(ctx context.Context, id string) error
}
