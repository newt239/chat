package auth_test

import (
	"context"
	"testing"

	"github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/test/mocks"
	authuc "github.com/newt239/chat/internal/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthInteractor_Register(t *testing.T) {
	tests := []struct {
		name           string
		input          authuc.RegisterInput
		setupMocks     func(*mocks.MockUserRepository, *mocks.MockSessionRepository, *mocks.MockJWTService, *mocks.MockPasswordService)
		expectedError  error
		expectedOutput *authuc.AuthOutput
	}{
		{
			name: "正常なユーザー登録",
			input: authuc.RegisterInput{
				Email:       "test@example.com",
				Password:    "password123",
				DisplayName: "Test User",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// ユーザーが存在しないことを確認
				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, nil)

				// パスワードハッシュ化
				passwordService.On("HashPassword", "password123").Return("hashed_password", nil)

				// ユーザー作成
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

				// セッション作成
				sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Session")).Return(nil)

				// JWTトークン生成
				jwtService.On("GenerateToken", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return("access_token", nil).Twice()
			},
			expectedError: nil,
		},
		{
			name: "既存ユーザーの登録",
			input: authuc.RegisterInput{
				Email:       "existing@example.com",
				Password:    "password123",
				DisplayName: "Existing User",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// 既存ユーザーが存在することを確認
				existingUser := mocks.TestUser("user-id", "existing@example.com", "Existing User")
				userRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: errors.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			userRepo := &mocks.MockUserRepository{}
			sessionRepo := &mocks.MockSessionRepository{}
			jwtService := &mocks.MockJWTService{}
			passwordService := &mocks.MockPasswordService{}

			tt.setupMocks(userRepo, sessionRepo, jwtService, passwordService)

			// インターラクターの作成
			interactor := authuc.NewAuthInteractor(userRepo, sessionRepo, jwtService, passwordService)

			// テスト実行
			output, err := interactor.Register(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotEmpty(t, output.AccessToken)
				assert.NotEmpty(t, output.RefreshToken)
				assert.Equal(t, tt.input.Email, output.User.Email)
				assert.Equal(t, tt.input.DisplayName, output.User.DisplayName)
			}

			// モックの検証
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			jwtService.AssertExpectations(t)
			passwordService.AssertExpectations(t)
		})
	}
}

func TestAuthInteractor_Login(t *testing.T) {
	tests := []struct {
		name          string
		input         authuc.LoginInput
		setupMocks    func(*mocks.MockUserRepository, *mocks.MockSessionRepository, *mocks.MockJWTService, *mocks.MockPasswordService)
		expectedError error
	}{
		{
			name: "正常なログイン",
			input: authuc.LoginInput{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// ユーザー検索
				user := mocks.TestUser("user-id", "test@example.com", "Test User")
				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

				// パスワード検証
				passwordService.On("VerifyPassword", "password123", "hashed_password").Return(nil)

				// セッション作成
				sessionRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Session")).Return(nil)

				// JWTトークン生成
				jwtService.On("GenerateToken", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return("access_token", nil).Twice()
			},
			expectedError: nil,
		},
		{
			name: "存在しないユーザー",
			input: authuc.LoginInput{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// ユーザーが見つからない
				userRepo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)
			},
			expectedError: errors.ErrInvalidCredentials,
		},
		{
			name: "間違ったパスワード",
			input: authuc.LoginInput{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// ユーザー検索
				user := mocks.TestUser("user-id", "test@example.com", "Test User")
				userRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

				// パスワード検証失敗
				passwordService.On("VerifyPassword", "wrongpassword", "hashed_password").Return(errors.ErrInvalidCredentials)
			},
			expectedError: errors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			userRepo := &mocks.MockUserRepository{}
			sessionRepo := &mocks.MockSessionRepository{}
			jwtService := &mocks.MockJWTService{}
			passwordService := &mocks.MockPasswordService{}

			tt.setupMocks(userRepo, sessionRepo, jwtService, passwordService)

			// インターラクターの作成
			interactor := authuc.NewAuthInteractor(userRepo, sessionRepo, jwtService, passwordService)

			// テスト実行
			output, err := interactor.Login(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotEmpty(t, output.AccessToken)
				assert.NotEmpty(t, output.RefreshToken)
			}

			// モックの検証
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			jwtService.AssertExpectations(t)
			passwordService.AssertExpectations(t)
		})
	}
}

func TestAuthInteractor_RefreshToken(t *testing.T) {
	tests := []struct {
		name          string
		input         authuc.RefreshTokenInput
		setupMocks    func(*mocks.MockUserRepository, *mocks.MockSessionRepository, *mocks.MockJWTService, *mocks.MockPasswordService)
		expectedError error
	}{
		{
			name: "正常なトークンリフレッシュ",
			input: authuc.RefreshTokenInput{
				RefreshToken: "valid_refresh_token",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// セッション検索
				session := mocks.TestSession("session-id", "user-id", "valid_refresh_token")
				sessionRepo.On("FindByRefreshToken", mock.Anything, "valid_refresh_token").Return(session, nil)

				// ユーザー検索
				user := mocks.TestUser("user-id", "test@example.com", "Test User")
				userRepo.On("FindByID", mock.Anything, "user-id").Return(user, nil)

				// セッション更新
				sessionRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Session")).Return(nil)

				// JWTトークン生成
				jwtService.On("GenerateToken", mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return("new_access_token", nil).Twice()
			},
			expectedError: nil,
		},
		{
			name: "無効なリフレッシュトークン",
			input: authuc.RefreshTokenInput{
				RefreshToken: "invalid_refresh_token",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// セッションが見つからない
				sessionRepo.On("FindByRefreshToken", mock.Anything, "invalid_refresh_token").Return(nil, nil)
			},
			expectedError: errors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			userRepo := &mocks.MockUserRepository{}
			sessionRepo := &mocks.MockSessionRepository{}
			jwtService := &mocks.MockJWTService{}
			passwordService := &mocks.MockPasswordService{}

			tt.setupMocks(userRepo, sessionRepo, jwtService, passwordService)

			// インターラクターの作成
			interactor := authuc.NewAuthInteractor(userRepo, sessionRepo, jwtService, passwordService)

			// テスト実行
			output, err := interactor.RefreshToken(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotEmpty(t, output.AccessToken)
				assert.NotEmpty(t, output.RefreshToken)
			}

			// モックの検証
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			jwtService.AssertExpectations(t)
			passwordService.AssertExpectations(t)
		})
	}
}

func TestAuthInteractor_Logout(t *testing.T) {
	tests := []struct {
		name          string
		input         authuc.LogoutInput
		setupMocks    func(*mocks.MockUserRepository, *mocks.MockSessionRepository, *mocks.MockJWTService, *mocks.MockPasswordService)
		expectedError error
	}{
		{
			name: "正常なログアウト",
			input: authuc.LogoutInput{
				UserID:       "user-id",
				RefreshToken: "refresh_token",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// セッション検索
				session := mocks.TestSession("session-id", "user-id", "refresh_token")
				sessionRepo.On("FindByRefreshToken", mock.Anything, "refresh_token").Return(session, nil)

				// セッション削除
				sessionRepo.On("Delete", mock.Anything, "session-id").Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "無効なリフレッシュトークン",
			input: authuc.LogoutInput{
				UserID:       "user-id",
				RefreshToken: "invalid_refresh_token",
			},
			setupMocks: func(userRepo *mocks.MockUserRepository, sessionRepo *mocks.MockSessionRepository, jwtService *mocks.MockJWTService, passwordService *mocks.MockPasswordService) {
				// セッションが見つからない
				sessionRepo.On("FindByRefreshToken", mock.Anything, "invalid_refresh_token").Return(nil, nil)
			},
			expectedError: errors.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの準備
			userRepo := &mocks.MockUserRepository{}
			sessionRepo := &mocks.MockSessionRepository{}
			jwtService := &mocks.MockJWTService{}
			passwordService := &mocks.MockPasswordService{}

			tt.setupMocks(userRepo, sessionRepo, jwtService, passwordService)

			// インターラクターの作成
			interactor := authuc.NewAuthInteractor(userRepo, sessionRepo, jwtService, passwordService)

			// テスト実行
			output, err := interactor.Logout(context.Background(), tt.input)

			// 結果の検証
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
			}

			// モックの検証
			userRepo.AssertExpectations(t)
			sessionRepo.AssertExpectations(t)
			jwtService.AssertExpectations(t)
			passwordService.AssertExpectations(t)
		})
	}
}
