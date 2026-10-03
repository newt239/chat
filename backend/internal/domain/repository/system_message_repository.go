package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type SystemMessageRepository interface {
	Create(ctx context.Context, msg *entity.SystemMessage) error
	FindByChannelIDs(ctx context.Context, channelIDs []string, limit int, since *time.Time, until *time.Time, ascending bool) ([]*entity.SystemMessage, error)
}
