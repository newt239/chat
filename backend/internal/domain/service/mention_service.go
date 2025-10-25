package service

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// MentionService defines the interface for mention operations
type MentionService interface {
	ExtractUserMentions(ctx context.Context, body, workspaceID string) ([]*entity.MessageUserMention, error)
	ExtractGroupMentions(ctx context.Context, body, workspaceID string) ([]*entity.MessageGroupMention, error)
}
