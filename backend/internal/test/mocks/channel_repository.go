package mocks

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
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

func (m *MockChannelRepository) SearchAccessibleChannels(ctx context.Context, workspaceID string, userID string, query string, limit int, offset int) ([]*entity.Channel, int, error) {
	args := m.Called(ctx, workspaceID, userID, query, limit, offset)
	var channels []*entity.Channel
	if result := args.Get(0); result != nil {
		channels = result.([]*entity.Channel)
	}
	total, _ := args.Get(1).(int)
	return channels, total, args.Error(2)
}

func (m *MockChannelRepository) FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error) {
	args := m.Called(ctx, workspaceID, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string, name string) (*entity.Channel, error) {
	args := m.Called(ctx, workspaceID, creatorID, memberIDs, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Channel), args.Error(1)
}

func (m *MockChannelRepository) FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
	args := m.Called(ctx, workspaceID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Channel), args.Error(1)
}

type MockChannelMemberRepository struct {
	mock.Mock
}

func NewMockChannelMemberRepository(t interface{}) *MockChannelMemberRepository {
	return &MockChannelMemberRepository{}
}

func (m *MockChannelMemberRepository) AddMember(ctx context.Context, member *entity.ChannelMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockChannelMemberRepository) RemoveMember(ctx context.Context, channelID string, userID string) error {
	args := m.Called(ctx, channelID, userID)
	return args.Error(0)
}

func (m *MockChannelMemberRepository) FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error) {
	args := m.Called(ctx, channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.ChannelMember), args.Error(1)
}

func (m *MockChannelMemberRepository) IsMember(ctx context.Context, channelID string, userID string) (bool, error) {
	args := m.Called(ctx, channelID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockChannelMemberRepository) UpdateMemberRole(ctx context.Context, channelID string, userID string, role entity.ChannelRole) error {
	args := m.Called(ctx, channelID, userID, role)
	return args.Error(0)
}

func (m *MockChannelMemberRepository) CountAdmins(ctx context.Context, channelID string) (int, error) {
	args := m.Called(ctx, channelID)
	return args.Int(0), args.Error(1)
}
