package repository

import (
	"context"

	"github.com/example/chat/internal/domain/entity"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error
}
