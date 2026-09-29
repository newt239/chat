package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type MessageRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Message, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Message, error)
	FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.Message, error)
	FindThreadReplies(ctx context.Context, parentID string) ([]*entity.Message, error)
	SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error
	Create(ctx context.Context, message *entity.Message) error
	Update(ctx context.Context, message *entity.Message) error
	AddReaction(ctx context.Context, reaction *entity.MessageReaction) error
	RemoveReaction(ctx context.Context, messageID string, userID string, emoji string) error
	FindReactions(ctx context.Context, messageID string) ([]*entity.MessageReaction, error)
	FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error)
	AddUserMention(ctx context.Context, mention *entity.MessageUserMention) error
	AddGroupMention(ctx context.Context, mention *entity.MessageGroupMention) error
	FindSearchScope(ctx context.Context, workspaceID string, userID string) (*MessageSearchScope, error)
	// FindSearchDocuments は指定したメッセージのうち削除されていないものの検索用文書を返します
	FindSearchDocuments(ctx context.Context, messageIDs []string) ([]MessageSearchDocument, error)
	// FindSearchDocumentsAfter は ID が afterID より大きい削除されていないメッセージの検索用文書を ID 順に limit 件返します
	FindSearchDocumentsAfter(ctx context.Context, afterID string, limit int) ([]MessageSearchDocument, error)
	FindMentions(ctx context.Context, input FindMentionsInput) ([]*entity.Message, error)
}
