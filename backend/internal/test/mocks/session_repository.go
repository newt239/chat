package mocks

import (
	"context"
	"time"

	"github.com/example/chat/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockSessionRepository はテスト用のSessionRepositoryモックです
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id string) (*entity.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Session), args.Error(1)
}

func (m *MockSessionRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (*entity.Session, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Session), args.Error(1)
}

func (m *MockSessionRepository) Create(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Update(ctx context.Context, session *entity.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*entity.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Session), args.Error(1)
}

func (m *MockSessionRepository) Revoke(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSessionRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// TestSession はテスト用のセッションエンティティを作成するヘルパー関数です
func TestSession(id, userID, refreshToken string) *entity.Session {
	now := time.Now()
	return &entity.Session{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: refreshToken,
		ExpiresAt:        now.Add(24 * time.Hour),
		CreatedAt:        now,
	}
}
