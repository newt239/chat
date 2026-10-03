package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type PushTokenRepository interface {
	// Upsert はトークンを登録します。別のユーザーに登録済みなら付け替えます
	Upsert(ctx context.Context, token *entity.PushToken) error
	Delete(ctx context.Context, userID string, token string) error
	// DeleteTokens は無効になったトークンを持ち主によらず削除します
	DeleteTokens(ctx context.Context, tokens []string) error
	FindByUserIDs(ctx context.Context, userIDs []string) ([]*entity.PushToken, error)
}
