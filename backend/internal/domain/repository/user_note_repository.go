package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type UserNoteRepository interface {
	Find(ctx context.Context, ownerID string, targetID string) (*entity.UserNote, error)
	// FindNicknames は ownerID が設定したニックネームを対象ユーザー ID ごとに返します
	FindNicknames(ctx context.Context, ownerID string) (map[string]string, error)
	Upsert(ctx context.Context, note *entity.UserNote) error
	Delete(ctx context.Context, ownerID string, targetID string) error
}
