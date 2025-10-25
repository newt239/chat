package workspace_test

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/test/mocks"
	workspaceuc "github.com/newt239/chat/internal/usecase/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkspaceInteractor_CreateWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		input         workspaceuc.CreateWorkspaceInput
		setupMocks    func(*mocks.MockWorkspaceRepository, *mocks.MockUserRepository)
		expectedError error
	}{
		{
			name: "正常なワークスペース作成",
			input: workspaceuc.CreateWorkspaceInput{
				Name:        "Test Workspace",
				Description: stringPtr("Test Description"),
				CreatedBy:   "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// ワークスペース作成
				workspaceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Workspace")).Return(nil)

				// メンバー追加
				workspaceRepo.On("AddMember", mock.Anything, mock.AnythingOfType("*entity.WorkspaceMember")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "ワークスペース作成失敗",
			input: workspaceuc.CreateWorkspaceInput{
				Name:        "Test Workspace",
				Description: stringPtr("Test Description"),
				CreatedBy:   "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// ワークスペース作成失敗
				workspaceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Workspace")).Return(errors.ErrInternal)
			},
			expectedError: errors.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			workspaceRepo := &mocks.MockWorkspaceRepository{}
			userRepo := &mocks.MockUserRepository{}

			tt.setupMocks(workspaceRepo, userRepo)

			// インターラクターの作成
			interactor := workspaceuc.NewWorkspaceInteractor(workspaceRepo, userRepo)

			// テスト実行
			output, err := interactor.CreateWorkspace(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.input.Name, output.Workspace.Name)
				assert.Equal(t, tt.input.Description, output.Workspace.Description)
			}

			// モックの検証
			workspaceRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestWorkspaceInteractor_GetWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		input         workspaceuc.GetWorkspaceInput
		setupMocks    func(*mocks.MockWorkspaceRepository, *mocks.MockUserRepository)
		expectedError error
	}{
		{
			name: "正常なワークスペース取得",
			input: workspaceuc.GetWorkspaceInput{
				ID:     "workspace-id",
				UserID: "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// メンバーシップ確認
				member := &entity.WorkspaceMember{
					WorkspaceID: "workspace-id",
					UserID:      "user-id",
					Role:        entity.WorkspaceRoleOwner,
				}
				workspaceRepo.On("FindMember", mock.Anything, "workspace-id", "user-id").Return(member, nil)

				// ワークスペース検索
				workspace := mocks.TestWorkspace("workspace-id", "Test Workspace", "Test Description")
				workspaceRepo.On("FindByID", mock.Anything, "workspace-id").Return(workspace, nil)
			},
			expectedError: nil,
		},
		{
			name: "権限なし",
			input: workspaceuc.GetWorkspaceInput{
				ID:     "workspace-id",
				UserID: "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// メンバーシップなし
				workspaceRepo.On("FindMember", mock.Anything, "workspace-id", "user-id").Return(nil, nil)
			},
			expectedError: workspaceuc.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			workspaceRepo := &mocks.MockWorkspaceRepository{}
			userRepo := &mocks.MockUserRepository{}

			tt.setupMocks(workspaceRepo, userRepo)

			// インターラクターの作成
			interactor := workspaceuc.NewWorkspaceInteractor(workspaceRepo, userRepo)

			// テスト実行
			output, err := interactor.GetWorkspace(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.input.ID, output.Workspace.ID)
			}

			// モックの検証
			workspaceRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestWorkspaceInteractor_GetWorkspacesByUserID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*mocks.MockWorkspaceRepository, *mocks.MockUserRepository)
		expectedError error
		expectedCount int
	}{
		{
			name:   "正常なワークスペース一覧取得",
			userID: "user-id",
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// ユーザーのワークスペース一覧取得
				desc1 := "Description 1"
				desc2 := "Description 2"
				workspaces := []*entity.Workspace{
					mocks.TestWorkspace("workspace-1", "Workspace 1", desc1),
					mocks.TestWorkspace("workspace-2", "Workspace 2", desc2),
				}
				workspaceRepo.On("FindByUserID", mock.Anything, "user-id").Return(workspaces, nil)

				// メンバー情報を返す
				member1 := &entity.WorkspaceMember{
					WorkspaceID: "workspace-1",
					UserID:      "user-id",
					Role:        entity.WorkspaceRoleOwner,
				}
				member2 := &entity.WorkspaceMember{
					WorkspaceID: "workspace-2",
					UserID:      "user-id",
					Role:        entity.WorkspaceRoleMember,
				}
				workspaceRepo.On("FindMember", mock.Anything, "workspace-1", "user-id").Return(member1, nil)
				workspaceRepo.On("FindMember", mock.Anything, "workspace-2", "user-id").Return(member2, nil)
			},
			expectedError: nil,
			expectedCount: 2,
		},
		{
			name:   "ワークスペースが存在しない",
			userID: "user-id",
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// 空のリストを返す
				workspaceRepo.On("FindByUserID", mock.Anything, "user-id").Return([]*entity.Workspace{}, nil)
			},
			expectedError: nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			workspaceRepo := &mocks.MockWorkspaceRepository{}
			userRepo := &mocks.MockUserRepository{}

			tt.setupMocks(workspaceRepo, userRepo)

			// インターラクターの作成
			interactor := workspaceuc.NewWorkspaceInteractor(workspaceRepo, userRepo)

			// テスト実行
			output, err := interactor.GetWorkspacesByUserID(context.Background(), tt.userID)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Len(t, output.Workspaces, tt.expectedCount)
			}

			// モックの検証
			workspaceRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestWorkspaceInteractor_UpdateWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		input         workspaceuc.UpdateWorkspaceInput
		setupMocks    func(*mocks.MockWorkspaceRepository, *mocks.MockUserRepository)
		expectedError error
	}{
		{
			name: "正常なワークスペース更新",
			input: workspaceuc.UpdateWorkspaceInput{
				ID:          "workspace-id",
				Name:        stringPtr("Updated Workspace"),
				Description: stringPtr("Updated Description"),
				UserID:      "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// メンバーシップ確認
				member := &entity.WorkspaceMember{
					WorkspaceID: "workspace-id",
					UserID:      "user-id",
					Role:        entity.WorkspaceRoleOwner,
				}
				workspaceRepo.On("FindMember", mock.Anything, "workspace-id", "user-id").Return(member, nil)

				// ワークスペース検索
				desc := "Original Description"
				workspace := mocks.TestWorkspace("workspace-id", "Original Workspace", desc)
				workspaceRepo.On("FindByID", mock.Anything, "workspace-id").Return(workspace, nil)

				// ワークスペース更新
				workspaceRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Workspace")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "権限なし",
			input: workspaceuc.UpdateWorkspaceInput{
				ID:     "workspace-id",
				Name:   stringPtr("Updated Workspace"),
				UserID: "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// メンバーシップなし
				workspaceRepo.On("FindMember", mock.Anything, "workspace-id", "user-id").Return(nil, nil)
			},
			expectedError: workspaceuc.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			workspaceRepo := &mocks.MockWorkspaceRepository{}
			userRepo := &mocks.MockUserRepository{}

			tt.setupMocks(workspaceRepo, userRepo)

			// インターラクターの作成
			interactor := workspaceuc.NewWorkspaceInteractor(workspaceRepo, userRepo)

			// テスト実行
			output, err := interactor.UpdateWorkspace(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				if tt.input.Name != nil {
					assert.Equal(t, *tt.input.Name, output.Workspace.Name)
				}
				if tt.input.Description != nil {
					assert.Equal(t, *tt.input.Description, *output.Workspace.Description)
				}
			}

			// モックの検証
			workspaceRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

func TestWorkspaceInteractor_DeleteWorkspace(t *testing.T) {
	tests := []struct {
		name          string
		input         workspaceuc.DeleteWorkspaceInput
		setupMocks    func(*mocks.MockWorkspaceRepository, *mocks.MockUserRepository)
		expectedError error
	}{
		{
			name: "正常なワークスペース削除",
			input: workspaceuc.DeleteWorkspaceInput{
				ID:     "workspace-id",
				UserID: "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// メンバーシップ確認
				member := &entity.WorkspaceMember{
					WorkspaceID: "workspace-id",
					UserID:      "user-id",
					Role:        entity.WorkspaceRoleOwner,
				}
				workspaceRepo.On("FindMember", mock.Anything, "workspace-id", "user-id").Return(member, nil)

				// ワークスペース削除
				workspaceRepo.On("Delete", mock.Anything, "workspace-id").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "権限なし",
			input: workspaceuc.DeleteWorkspaceInput{
				ID:     "workspace-id",
				UserID: "user-id",
			},
			setupMocks: func(workspaceRepo *mocks.MockWorkspaceRepository, userRepo *mocks.MockUserRepository) {
				// メンバーシップなし
				workspaceRepo.On("FindMember", mock.Anything, "workspace-id", "user-id").Return(nil, nil)
			},
			expectedError: workspaceuc.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			workspaceRepo := &mocks.MockWorkspaceRepository{}
			userRepo := &mocks.MockUserRepository{}

			tt.setupMocks(workspaceRepo, userRepo)

			// インターラクターの作成
			interactor := workspaceuc.NewWorkspaceInteractor(workspaceRepo, userRepo)

			// テスト実行
			output, err := interactor.DeleteWorkspace(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.True(t, output.Success)
			}

			// モックの検証
			workspaceRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
		})
	}
}

// ヘルパー関数
func stringPtr(s string) *string {
	return &s
}
