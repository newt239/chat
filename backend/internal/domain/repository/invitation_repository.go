package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.Invitation) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.Invitation, error)
	// FindPendingByEmail は未受諾で期限内の招待をワークスペースをまたいで返します
	FindPendingByEmail(ctx context.Context, email string, now time.Time) ([]*entity.Invitation, error)
	// FindPendingByWorkspaceID は未受諾で期限内の招待を作成日時の降順で返します
	FindPendingByWorkspaceID(ctx context.Context, workspaceID string, now time.Time) ([]*entity.Invitation, error)
	MarkAccepted(ctx context.Context, id string, acceptedAt time.Time) error
	// Delete は指定したワークスペースの招待だけを削除し、見つからなければ ErrNotFound を返します
	Delete(ctx context.Context, workspaceID, id string) error
}
