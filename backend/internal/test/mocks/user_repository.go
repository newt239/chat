package mocks

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository はテスト用のUserRepositoryモックです
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestUser はテスト用のユーザーエンティティを作成するヘルパー関数です
func TestUser(id, email, displayName string) *entity.User {
	now := time.Now()
	return &entity.User{
		ID:           id,
		Email:        email,
		PasswordHash: "hashed_password",
		DisplayName:  displayName,
		AvatarURL:    nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
