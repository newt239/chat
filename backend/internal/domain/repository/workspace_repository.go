package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type WorkspaceRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Workspace, error)
	FindByUserID(ctx context.Context, userID string) ([]*entity.Workspace, error)
	Create(ctx context.Context, workspace *entity.Workspace) error
	Update(ctx context.Context, workspace *entity.Workspace) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *entity.WorkspaceMember) error
	UpdateMemberRole(ctx context.Context, workspaceID string, userID string, role entity.WorkspaceRole) error
	RemoveMember(ctx context.Context, workspaceID string, userID string) error
	FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error)
	FindMember(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error)
}
