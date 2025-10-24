package websocket

import (
	"log"
	"net/http"

	authuc "github.com/example/chat/internal/usecase/auth"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: 本番環境では適切なオリジンチェックを実装
		return true
	},
}

// Handler はWebSocketハンドラーを返します
func NewHandler(hub *Hub, jwtService authuc.JWTService) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 認証チェック
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "authorization header required")
		}

		// JWTトークンの検証
		token := authHeader
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := jwtService.VerifyToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
		}

		// WebSocket接続のアップグレード
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return err
		}

		// クライアントを作成してハブに登録
		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan []byte, 256),
			userID: claims.UserID,
		}

		client.hub.register <- client

		// ゴルーチンを開始
		go client.writePump()
		go client.readPump()

		return nil
	}
}
