package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/domain/repository"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

// MessageUseCase はメッセージユースケースのインターフェースです
type MessageUseCase interface {
	// メッセージ関連の操作（必要に応じて定義）
}

// ReadStateUseCase は既読状態ユースケースのインターフェースです
type ReadStateUseCase interface {
	// 既読状態関連の操作（必要に応じて定義）
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 本番環境では適切なオリジンチェックを実装
		return true
	},
}

// Handler はWebSocketハンドラーを返します
func Handler(hub *Hub, jwtService authuc.JWTService, workspaceRepo repository.WorkspaceRepository, messageUseCase MessageUseCase, readStateUseCase ReadStateUseCase) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 認証トークンの取得
		// WebSocketではAuthorizationヘッダーを設定できないため、クエリパラメータからも取得を試みる
		var token string
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != "" {
			token = authHeader
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}
		} else {
			// クエリパラメータからトークンを取得
			token = c.QueryParam("token")
			if token == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication token required")
			}
		}

		// JWTトークンの検証
		claims, err := jwtService.VerifyToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
		}

		// WorkspaceIDの取得
		workspaceID := c.QueryParam("workspaceId")
		if workspaceID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "workspaceId query parameter required")
		}

		// Workspace所属確認
		ctx := c.Request().Context()
		member, err := workspaceRepo.FindMember(ctx, workspaceID, claims.UserID)
		if err != nil || member == nil {
			return echo.NewHTTPError(http.StatusForbidden, "user is not a member of this workspace")
		}

		// WebSocket接続のアップグレード
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return err
		}

		// クライアントを作成してハブに登録
		client := &Client{
			hub:                hub,
			conn:               conn,
			send:               make(chan []byte, 256),
			userID:             claims.UserID,
			workspaceID:        workspaceID,
			subscribedChannels: make(map[string]bool),
			messageUseCase:     messageUseCase,
			readStateUseCase:   readStateUseCase,
		}

		client.hub.register <- client

		// ゴルーチンを開始
		go client.writePump()
		go client.readPump()

		return nil
	}
}
