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
	SearchMembers(ctx context.Context, workspaceID string, query string, limit int, offset int) ([]*entity.WorkspaceMember, int, error)

    // 新規メソッド
    FindAllPublic(ctx context.Context) ([]*entity.Workspace, error)
    CountMembers(ctx context.Context, workspaceID string) (int, error)
    ExistsByID(ctx context.Context, id string) (bool, error)
}
