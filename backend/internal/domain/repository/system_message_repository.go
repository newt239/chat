package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// SystemMessageRepository はシステムメッセージの永続化を扱います
type SystemMessageRepository interface {
	Create(ctx context.Context, msg *entity.SystemMessage) error
	FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.SystemMessage, error)
}
