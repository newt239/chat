package websocket

import (
	"log"
	"net/http"
	"slices"

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

// newUpgrader は許可オリジンのみ受け付ける Upgrader を作ります
func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// 同一オリジンやブラウザ以外からの接続は Origin を持たない
			if origin == "" {
				return true
			}
			return slices.Contains(allowedOrigins, "*") || slices.Contains(allowedOrigins, origin)
		},
	}
}

// Handler はWebSocketハンドラーを返します
func Handler(hub *Hub, jwtService authuc.JWTService, workspaceRepo repository.WorkspaceRepository, messageUseCase MessageUseCase, readStateUseCase ReadStateUseCase, allowedOrigins []string) echo.HandlerFunc {
	upgrader := newUpgrader(allowedOrigins)

	return func(c echo.Context) error {
		log.Printf("[WebSocket] 接続リクエスト受信: RemoteAddr=%s", c.Request().RemoteAddr)

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
				log.Printf("[WebSocket] 認証トークンが指定されていません: RemoteAddr=%s", c.Request().RemoteAddr)
				return echo.NewHTTPError(http.StatusUnauthorized, "認証トークンが指定されていません")
			}
		}

		// JWTトークンの検証
		claims, err := jwtService.VerifyToken(token)
		if err != nil {
			log.Printf("[WebSocket] トークン検証失敗: err=%v RemoteAddr=%s", err, c.Request().RemoteAddr)
			return echo.NewHTTPError(http.StatusUnauthorized, "トークンが無効または期限切れです")
		}

		// WorkspaceIDの取得
		workspaceID := c.QueryParam("workspaceId")
		if workspaceID == "" {
			log.Printf("[WebSocket] workspaceIdが指定されていません: userID=%s RemoteAddr=%s", claims.UserID, c.Request().RemoteAddr)
			return echo.NewHTTPError(http.StatusBadRequest, "workspaceIdクエリパラメータは必須です")
		}

		// Workspace所属確認
		ctx := c.Request().Context()
		member, err := workspaceRepo.FindMember(ctx, workspaceID, claims.UserID)
		if err != nil {
			log.Printf("[WebSocket] FindMember error: userID=%s workspaceID=%s err=%v", claims.UserID, workspaceID, err)
			return echo.NewHTTPError(http.StatusForbidden, "ユーザーはこのワークスペースのメンバーではありません")
		}
		if member == nil {
			log.Printf("[WebSocket] Member not found: userID=%s workspaceID=%s", claims.UserID, workspaceID)
			return echo.NewHTTPError(http.StatusForbidden, "ユーザーはこのワークスペースのメンバーではありません")
		}

		log.Printf("[WebSocket] 認証成功、アップグレード開始: userID=%s workspaceID=%s", claims.UserID, workspaceID)

		// WebSocket接続のアップグレード
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("[WebSocket] アップグレード失敗: userID=%s workspaceID=%s err=%v", claims.UserID, workspaceID, err)
			return err
		}

		log.Printf("[WebSocket] アップグレード成功: userID=%s workspaceID=%s", claims.UserID, workspaceID)

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
