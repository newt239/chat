package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageUserMentionRepository interface {
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageUserMention, error)
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, error)
	FindByUserID(ctx context.Context, userID string, limit int, since *time.Time) ([]*entity.MessageUserMention, error)
	Create(ctx context.Context, mention *entity.MessageUserMention) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}

type MessageGroupMentionRepository interface {
	FindByMessageID(ctx context.Context, messageID string) ([]*entity.MessageGroupMention, error)
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageGroupMention, error)
	FindByGroupID(ctx context.Context, groupID string, limit int, since *time.Time) ([]*entity.MessageGroupMention, error)
	Create(ctx context.Context, mention *entity.MessageGroupMention) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}
