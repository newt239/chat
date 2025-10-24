package mocks

import (
	"context"

	"github.com/example/chat/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockChannelRepository struct {
	mock.Mock
}

func NewMockChannelRepository(t interface{}) *MockChannelRepository {
	return &MockChannelRepository{}
}

func (m *MockChannelRepository) FindByID(ctx context.Context, id string) (*entity.Channel, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
	args := m.Called(ctx, workspaceID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) Create(ctx context.Context, channel *entity.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) Update(ctx context.Context, channel *entity.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *MockChannelRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockChannelRepository) AddMember(ctx context.Context, member *entity.ChannelMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockChannelRepository) RemoveMember(ctx context.Context, channelID string, userID string) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

func (m *MockChannelRepository) FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error) {
	args := m.Called(ctx, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ChannelMember), args.Error(1)
}

func (m *MockChannelRepository) IsMember(ctx context.Context, channelID string, userID string) (bool, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.Error(1)
}
