package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageMentionRepository interface {
	FindByMessageIDs(ctx context.Context, messageIDs []string) ([]*entity.MessageUserMention, []*entity.MessageGroupMention, error)
	Create(ctx context.Context, users []*entity.MessageUserMention, groups []*entity.MessageGroupMention) error
	DeleteByMessageID(ctx context.Context, messageID string) error
}
