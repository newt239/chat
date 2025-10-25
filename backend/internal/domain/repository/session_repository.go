package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type SessionRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Session, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error)
	Create(ctx context.Context, session *entity.Session) error
	Revoke(ctx context.Context, id string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}
