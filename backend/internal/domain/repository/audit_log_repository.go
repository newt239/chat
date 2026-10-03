package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	// List は新しい順に返します
	List(ctx context.Context, filter entity.AuditLogFilter, limit, offset int) ([]*entity.AuditLog, error)
}
