package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ReminderRepository interface {
	Create(ctx context.Context, reminder *entity.Reminder) error
	// ClaimDue は staleBefore より前から送信中のものを失敗にしてから、期限の来たリマインダーを最大 limit 件、他のワーカーと重ならないよう送信中にして返します
	ClaimDue(ctx context.Context, now, staleBefore time.Time, limit int) ([]*entity.Reminder, error)
	MarkSent(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
}
