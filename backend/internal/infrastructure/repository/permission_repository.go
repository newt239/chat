package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/workspacepermission"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type permissionRepository struct {
	client *ent.Client
}

func NewPermissionRepository(client *ent.Client) domainrepository.PermissionRepository {
	return &permissionRepository{client: client}
}

func (r *permissionRepository) FindOverrides(ctx context.Context, workspaceID string) ([]entity.PermissionOverride, error) {
	client := transaction.ResolveClient(ctx, r.client)
	rows, err := client.WorkspacePermission.Query().
		Where(workspacepermission.WorkspaceID(workspaceID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]entity.PermissionOverride, 0, len(rows))
	for _, row := range rows {
		result = append(result, entity.PermissionOverride{
			Role:       entity.WorkspaceRole(row.Role),
			Permission: entity.Permission(row.Permission),
			Allowed:    row.Allowed,
		})
	}
	return result, nil
}

func (r *permissionRepository) Upsert(ctx context.Context, workspaceID string, override entity.PermissionOverride) error {
	client := transaction.ResolveClient(ctx, r.client)
	return client.WorkspacePermission.Create().
		SetWorkspaceID(workspaceID).
		SetRole(string(override.Role)).
		SetPermission(string(override.Permission)).
		SetAllowed(override.Allowed).
		OnConflictColumns(
			workspacepermission.FieldWorkspaceID,
			workspacepermission.FieldRole,
			workspacepermission.FieldPermission,
		).
		UpdateAllowed().
		UpdateUpdatedAt().
		Exec(ctx)
}
