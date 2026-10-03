package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageUserMentionRepository interface {
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, error)
	CreateBulk(ctx context.Context, mentions []*entity.MessageUserMention) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}

type MessageGroupMentionRepository interface {
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageGroupMention, error)
	CreateBulk(ctx context.Context, mentions []*entity.MessageGroupMention) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}
