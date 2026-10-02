package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type WorkspaceRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Workspace, error)
	FindByUserID(ctx context.Context, userID string) ([]*entity.Workspace, error)
	// Create は ID が使われていれば ErrWorkspaceIDExists を返します
	Create(ctx context.Context, workspace *entity.Workspace) error
	Update(ctx context.Context, workspace *entity.Workspace) error
	Delete(ctx context.Context, id string) error
	AddMember(ctx context.Context, member *entity.WorkspaceMember) error
	UpdateMemberRole(ctx context.Context, workspaceID string, userID string, role entity.WorkspaceRole) error
	RemoveMember(ctx context.Context, workspaceID string, userID string) error
	FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error)
	// FindMember は停止中のメンバーを返しません。停止中も含める場合は FindMemberIncludingSuspended を使います
	FindMember(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error)
	FindMemberIncludingSuspended(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error)
	SetMemberSuspended(ctx context.Context, workspaceID string, userID string, suspendedAt *time.Time) error
	SearchMembers(ctx context.Context, workspaceID string, query string, limit int, offset int) ([]*entity.WorkspaceMember, int, error)
	FindAllPublic(ctx context.Context) ([]*entity.Workspace, error)
	CountMembersBatch(ctx context.Context, workspaceIDs []string) (map[string]int, error)
	// FindMembershipsByUserID は停止されずに参加しているワークスペースのメンバー情報を返します
	FindMembershipsByUserID(ctx context.Context, userID string) ([]*entity.WorkspaceMember, error)
}
