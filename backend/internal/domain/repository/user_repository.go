package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	// FindByIDs は見つかったユーザーを ID ごとに返します
	FindByIDs(ctx context.Context, ids []string) (map[string]*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByGoogleSub(ctx context.Context, sub string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	// Delete は本人だけのデータを消し、行を匿名化して退会済みにします
	Delete(ctx context.Context, id string) error
}
