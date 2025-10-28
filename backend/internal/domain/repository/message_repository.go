package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Message, error)
	FindByChannelID(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error)
	FindByChannelIDIncludingDeleted(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error)
	FindThreadReplies(ctx context.Context, parentID string) ([]*entity.Message, error)
	FindThreadRepliesIncludingDeleted(ctx context.Context, parentID string) ([]*entity.Message, error)
	SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error
	Create(ctx context.Context, message *entity.Message) error
	Update(ctx context.Context, message *entity.Message) error
	Delete(ctx context.Context, id string) error
	AddReaction(ctx context.Context, reaction *entity.MessageReaction) error
	RemoveReaction(ctx context.Context, messageID string, userID string, emoji string) error
	FindReactions(ctx context.Context, messageID string) ([]*entity.MessageReaction, error)
	FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error)
	AddUserMention(ctx context.Context, mention *entity.MessageUserMention) error
	AddGroupMention(ctx context.Context, mention *entity.MessageGroupMention) error
}
