package mocks

import (
	"context"
	"time"

	"github.com/example/chat/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

// MockWorkspaceRepository はテスト用のWorkspaceRepositoryモックです
type MockWorkspaceRepository struct {
	mock.Mock
}

func NewMockWorkspaceRepository(t interface{}) *MockWorkspaceRepository {
	return &MockWorkspaceRepository{}
}

func (m *MockWorkspaceRepository) FindByID(ctx context.Context, id string) (*entity.Workspace, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Workspace, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Workspace), args.Error(1)
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	args := m.Called(ctx, workspace)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID string) error {
	args := m.Called(ctx, workspaceID, userID)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID string, role entity.WorkspaceRole) error {
	args := m.Called(ctx, workspaceID, userID, role)
	return args.Error(0)
}

func (m *MockWorkspaceRepository) FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.WorkspaceMember), args.Error(1)
}

func (m *MockWorkspaceRepository) FindMember(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error) {
	args := m.Called(ctx, workspaceID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.WorkspaceMember), args.Error(1)
}

// TestWorkspace はテスト用のワークスペースエンティティを作成するヘルパー関数です
func TestWorkspace(id, name, description string) *entity.Workspace {
	now := time.Now()
	return &entity.Workspace{
		ID:          id,
		Name:        name,
		Description: &description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
