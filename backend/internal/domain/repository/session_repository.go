package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type SessionRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Session, error)
	// FindActiveByTokenHash は失効・期限切れでないセッションだけを返します
	FindActiveByTokenHash(ctx context.Context, refreshTokenHash string) (*entity.Session, error)
	Create(ctx context.Context, session *entity.Session) error
	Revoke(ctx context.Context, id string) error
	// Rotate はリフレッシュ時にセッションを作り直さずトークンだけ差し替えます
	Rotate(ctx context.Context, id string, refreshTokenHash string, expiresAt time.Time) error
	// FindLatestByUserIDs はユーザーごとに最後にログインしたセッションを返します
	FindLatestByUserIDs(ctx context.Context, userIDs []string) (map[string]*entity.Session, error)
	RevokeAllByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}
