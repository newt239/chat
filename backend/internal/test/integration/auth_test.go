package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/infrastructure/auth"
	"github.com/newt239/chat/internal/infrastructure/config"
	"github.com/newt239/chat/internal/infrastructure/repository"
	httpv "github.com/newt239/chat/internal/interfaces/handler/http"
	"github.com/newt239/chat/internal/registry"
	"github.com/newt239/chat/internal/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthIntegration(t *testing.T) {
	// テスト用データベースのセットアップ
	testDB := integration.NewTestDB(t)
	defer testDB.Cleanup(t)

	// 設定の読み込み
	cfg, err := config.Load()
	require.NoError(t, err)

	// レジストリの作成
	reg := registry.NewRegistry(testDB.Client, cfg)

	// 認証ハンドラーの作成
	authHandler := reg.Interface().NewAuthHandler()

	// Echoアプリケーションのセットアップ
	e := echo.New()
	e.Validator = httpv.NewValidator()
	e.POST("/auth/register", authHandler.Register)
	e.POST("/auth/login", authHandler.Login)
	e.POST("/auth/refresh", authHandler.RefreshToken)
	e.POST("/auth/logout", authHandler.Logout)

	t.Run("ユーザー登録", func(t *testing.T) {
		// リクエストボディの準備
		reqBody := map[string]string{
			"email":        "newuser@example.com",
			"password":     "password123",
			"display_name": "New User",
		}
		reqJSON, _ := json.Marshal(reqBody)

		// リクエストの作成
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// リクエストの実行
		c := e.NewContext(req, rec)
		err := authHandler.Register(c)

		// 結果の検証
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// レスポンスの検証
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "accessToken")
		assert.Contains(t, response, "refreshToken")
		assert.Contains(t, response, "user")
	})

	t.Run("ユーザーログイン", func(t *testing.T) {
		// テスト用ユーザーの作成
		userRepo := repository.NewUserRepository(testDB.Client)
		passwordService := auth.NewPasswordService()

		hashedPassword, err := passwordService.HashPassword("password123")
		require.NoError(t, err)

		user := &entity.User{
			Email:        "testuser@example.com",
			PasswordHash: hashedPassword,
			DisplayName:  "Test User",
		}
		err = userRepo.Create(context.Background(), user)
		require.NoError(t, err)

		// リクエストボディの準備
		reqBody := map[string]string{
			"email":    "testuser@example.com",
			"password": "password123",
		}
		reqJSON, _ := json.Marshal(reqBody)

		// リクエストの作成
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// リクエストの実行
		c := e.NewContext(req, rec)
		err = authHandler.Login(c)

		// 結果の検証
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		// レスポンスの検証
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "accessToken")
		assert.Contains(t, response, "refreshToken")
		assert.Contains(t, response, "user")
	})

	t.Run("無効な認証情報でのログイン", func(t *testing.T) {
		// リクエストボディの準備
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "wrongpassword",
		}
		reqJSON, _ := json.Marshal(reqBody)

		// リクエストの作成
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqJSON))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// リクエストの実行
		c := e.NewContext(req, rec)
		err := authHandler.Login(c)

		// 結果の検証
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
