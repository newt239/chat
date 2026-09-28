package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type PermissionRepository interface {
	FindOverrides(ctx context.Context, workspaceID string) ([]entity.PermissionOverride, error)
	Upsert(ctx context.Context, workspaceID string, override entity.PermissionOverride) error
}
