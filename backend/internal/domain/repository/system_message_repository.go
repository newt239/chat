package repository

import (
    "context"
    "time"

    "github.com/newt239/chat/internal/domain/entity"
)

// SystemMessageRepository はシステムメッセージの永続化を扱います
type SystemMessageRepository interface {
    Create(ctx context.Context, msg *entity.SystemMessage) error
    FindByChannelID(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.SystemMessage, error)
}


