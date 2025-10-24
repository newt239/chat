package mocks

import (
	"context"

	"github.com/example/chat/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockBookmarkRepository struct {
	mock.Mock
}

func NewMockBookmarkRepository(t interface{}) *MockBookmarkRepository {
	return &MockBookmarkRepository{}
}

func (m *MockBookmarkRepository) AddBookmark(ctx context.Context, bookmark *entity.MessageBookmark) error {
	args := m.Called(ctx, bookmark)
	return args.Error(0)
}

func (m *MockBookmarkRepository) RemoveBookmark(ctx context.Context, userID, messageID string) error {
	args := m.Called(ctx, userID, messageID)
	return args.Error(0)
}

func (m *MockBookmarkRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.MessageBookmark, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.MessageBookmark), args.Error(1)
}

func (m *MockBookmarkRepository) IsBookmarked(ctx context.Context, userID, messageID string) (bool, error) {
	args := m.Called(ctx, userID, messageID)
	return args.Bool(0), args.Error(1)
}
