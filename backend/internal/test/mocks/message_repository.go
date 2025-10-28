package mocks

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockMessageRepository struct {
	mock.Mock
}

func NewMockMessageRepository(t interface{}) *MockMessageRepository {
	return &MockMessageRepository{}
}

func (m *MockMessageRepository) FindByID(ctx context.Context, id string) (*entity.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) FindByChannelID(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error) {
	args := m.Called(ctx, channelID, limit, since, until)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) FindThreadReplies(ctx context.Context, parentID string) ([]*entity.Message, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) Create(ctx context.Context, message *entity.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) Update(ctx context.Context, message *entity.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) AddReaction(ctx context.Context, reaction *entity.MessageReaction) error {
	args := m.Called(ctx, reaction)
	return args.Error(0)
}

func (m *MockMessageRepository) RemoveReaction(ctx context.Context, messageID string, userID string, emoji string) error {
	args := m.Called(ctx, messageID, userID, emoji)
	return args.Error(0)
}

func (m *MockMessageRepository) FindReactions(ctx context.Context, messageID string) ([]*entity.MessageReaction, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.MessageReaction), args.Error(1)
}

func (m *MockMessageRepository) FindReactionsByMessageIDs(ctx context.Context, messageIDs []string) (map[string][]*entity.MessageReaction, error) {
	args := m.Called(ctx, messageIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]*entity.MessageReaction), args.Error(1)
}

func (m *MockMessageRepository) FindByChannelIDIncludingDeleted(ctx context.Context, channelID string, limit int, since *time.Time, until *time.Time) ([]*entity.Message, error) {
	args := m.Called(ctx, channelID, limit, since, until)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) FindThreadRepliesIncludingDeleted(ctx context.Context, parentID string) ([]*entity.Message, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Message), args.Error(1)
}

func (m *MockMessageRepository) SoftDeleteByIDs(ctx context.Context, ids []string, deletedBy string) error {
	args := m.Called(ctx, ids, deletedBy)
	return args.Error(0)
}

func (m *MockMessageRepository) AddUserMention(ctx context.Context, mention *entity.MessageUserMention) error {
	args := m.Called(ctx, mention)
	return args.Error(0)
}

func (m *MockMessageRepository) AddGroupMention(ctx context.Context, mention *entity.MessageGroupMention) error {
	args := m.Called(ctx, mention)
	return args.Error(0)
}
