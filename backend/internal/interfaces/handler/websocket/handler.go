package websocket

import (
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	domerr "github.com/newt239/chat/internal/domain/errors"
	realtimeuc "github.com/newt239/chat/internal/usecase/realtime"
)

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

// Handler は RealtimeService で発行した 1 回限りのチケットで認証し、WebSocket に切り替えます
func Handler(hub *Hub, tickets *realtimeuc.Interactor, allowedOrigins []string) echo.HandlerFunc {
	upgrader := newUpgrader(allowedOrigins)

	return func(c echo.Context) error {
		ticket, err := tickets.ConsumeTicket(c.Request().Context(), c.QueryParam("ticket"))
		if errors.Is(err, domerr.ErrForbidden) {
			return echo.NewHTTPError(http.StatusForbidden, "ワークスペースのメンバーではありません")
		}
		if err != nil {
			if !errors.Is(err, domerr.ErrInvalidToken) {
				slog.Error("WebSocket のチケットを確認できません", "error", err)
			}
			return echo.NewHTTPError(http.StatusUnauthorized, "チケットが無効か期限切れです")
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			// Upgrade がエラーの応答を書き込み済み
			return nil
		}
		newClient(hub, conn, uuid.NewString(), ticket.UserID, ticket.SessionID, ticket.WorkspaceID).start()
		return nil
	}
}
