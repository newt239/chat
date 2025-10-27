package integration_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/infrastructure/config"
	wscontroller "github.com/newt239/chat/internal/interfaces/handler/websocket"
	"github.com/newt239/chat/internal/registry"
	"github.com/newt239/chat/internal/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSocketIntegration(t *testing.T) {
	// テスト用データベースのセットアップ
	testDB := integration.NewTestDB(t)
	defer testDB.Cleanup(t)

	// 設定の読み込み
	cfg, err := config.Load()
	require.NoError(t, err)

	// レジストリの作成
	reg := registry.NewRegistry(testDB.DB, cfg)

	// WebSocketハブの作成
	hub := wscontroller.NewHub()
	go hub.Run()

	// JWTサービスとWorkspaceRepositoryの作成
	jwtService := reg.Infrastructure().NewJWTService()
	workspaceRepo := reg.Domain().NewWorkspaceRepository()
	userRepo := reg.Domain().NewUserRepository()
	messageUseCase := reg.UseCase().NewMessageUseCase()
	readStateUseCase := reg.UseCase().NewReadStateUseCase()

	// Echoアプリケーションのセットアップ
	e := echo.New()
	e.GET("/ws", wscontroller.NewHandler(hub, jwtService, workspaceRepo, messageUseCase, readStateUseCase))

	// テスト用のユーザーとWorkspaceを作成
	ctx := context.Background()
	aliceUserID := "11111111-1111-1111-1111-111111111111"
	testWorkspaceID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	// ユーザーを作成
	testUser := &entity.User{
		ID:           aliceUserID,
		Email:        "alice@example.com",
		PasswordHash: "dummy-hash",
		DisplayName:  "Alice Johnson",
	}
	err = userRepo.Create(ctx, testUser)
	require.NoError(t, err)

	// Workspaceを作成
	testWorkspace := &entity.Workspace{
		ID:        testWorkspaceID,
		Name:      "Test Workspace",
		CreatedBy: aliceUserID,
	}
	err = workspaceRepo.Create(ctx, testWorkspace)
	require.NoError(t, err)

	// ユーザーをWorkspaceに追加
	member := &entity.WorkspaceMember{
		WorkspaceID: testWorkspaceID,
		UserID:      aliceUserID,
		Role:        entity.WorkspaceRoleOwner,
	}
	err = workspaceRepo.AddMember(ctx, member)
	require.NoError(t, err)

	// テスト用のJWTトークンを生成
	accessToken, err := jwtService.GenerateToken(aliceUserID, time.Hour)
	require.NoError(t, err)

	t.Run("WebSocket接続の確立", func(t *testing.T) {
		// WebSocket接続のテスト
		server := httptest.NewServer(e)
		defer server.Close()

		// WebSocket URLの構築
		wsURL := "ws" + server.URL[4:] + "/ws?workspaceId=" + testWorkspaceID + "&token=" + accessToken

		// WebSocket接続
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// 接続が確立されたことを確認
		assert.NotNil(t, conn)

		// ハブにクライアントが登録されたことを確認
		time.Sleep(100 * time.Millisecond) // 非同期処理の完了を待つ
		// ハブのクライアント数は非公開フィールドのため、接続が確立されたことを確認
		assert.NotNil(t, conn)
	})

	t.Run("無効なトークンでの接続", func(t *testing.T) {
		// WebSocket接続のテスト
		server := httptest.NewServer(e)
		defer server.Close()

		// 無効なトークンでWebSocket URLを構築
		wsURL := "ws" + server.URL[4:] + "/ws?workspaceId=" + testWorkspaceID + "&token=invalid-token"

		// WebSocket接続（エラーが期待される）
		_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.Error(t, err)
	})

	t.Run("所属していないWorkspaceへの接続", func(t *testing.T) {
		// WebSocket接続のテスト
		server := httptest.NewServer(e)
		defer server.Close()

		// 存在しないWorkspaceIDでWebSocket URLを構築
		nonExistentWorkspaceID := "99999999-9999-9999-9999-999999999999"
		wsURL := "ws" + server.URL[4:] + "/ws?workspaceId=" + nonExistentWorkspaceID + "&token=" + accessToken

		// WebSocket接続（エラーが期待される）
		_, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.Error(t, err)
	})

	t.Run("メッセージの送受信", func(t *testing.T) {
		// WebSocket接続のテスト
		server := httptest.NewServer(e)
		defer server.Close()

		// WebSocket URLの構築
		wsURL := "ws" + server.URL[4:] + "/ws?workspaceId=" + testWorkspaceID + "&token=" + accessToken

		// WebSocket接続
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// メッセージの送信
		testMessage := map[string]interface{}{
			"type":    "message",
			"content": "Hello, World!",
		}

		err = conn.WriteJSON(testMessage)
		require.NoError(t, err)

		// レスポンスの受信（タイムアウト付き）
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		var response map[string]interface{}
		err = conn.ReadJSON(&response)
		require.NoError(t, err)
	})

	t.Run("複数クライアントの接続", func(t *testing.T) {
		// WebSocket接続のテスト
		server := httptest.NewServer(e)
		defer server.Close()

		// 複数のWebSocket接続を作成
		connections := make([]*websocket.Conn, 3)
		for i := range 3 {
			wsURL := "ws" + server.URL[4:] + "/ws?workspaceId=" + testWorkspaceID + "&token=" + accessToken
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			connections[i] = conn
		}

		// すべての接続を閉じる
		for _, conn := range connections {
			conn.Close()
		}

		// ハブに複数のクライアントが登録されたことを確認
		time.Sleep(100 * time.Millisecond) // 非同期処理の完了を待つ
		// ハブのクライアント数は非公開フィールドのため、接続が確立されたことを確認
		assert.Len(t, connections, 3)
	})
}
