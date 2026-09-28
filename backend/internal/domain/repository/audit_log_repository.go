package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	// List は新しい順に返し、2 つ目の戻り値でページングを無視した総件数を返します
	List(ctx context.Context, filter entity.AuditLogFilter) ([]*entity.AuditLog, int, error)
}
