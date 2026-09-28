package http

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type RouterConfig struct {
	JWTService     authuc.JWTService
	AllowedOrigins []string

	WebSocketHub        *websocket.Hub
	WorkspaceRepository repository.WorkspaceRepository
	ChannelAccess       service.ChannelAccessService
	RPCHandler          http.Handler
	WebhookPoster       WebhookPoster
	// 開発用のローカルストレージを使うときだけ設定する
	StorageHandler http.Handler
}

func NewRouter(cfg RouterConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     append(connectcors.AllowedMethods(), http.MethodPut),
		AllowHeaders:     append(connectcors.AllowedHeaders(), echo.HeaderAuthorization),
		ExposeHeaders:    connectcors.ExposedHeaders(),
		AllowCredentials: true,
	}))

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ws", websocket.Handler(cfg.WebSocketHub, cfg.JWTService, cfg.WorkspaceRepository, cfg.ChannelAccess, cfg.AllowedOrigins))

	e.Any("/chat.v1.*", echo.WrapHandler(cfg.RPCHandler))

	// Webhook ごとに毎秒 1 回、瞬間的には 10 回まで受け付ける
	e.POST("/webhooks/:id/:token", webhookHandler(cfg.WebhookPoster, newRateLimiter(1, 10)))

	if cfg.StorageHandler != nil {
		e.Any("/storage/*", echo.WrapHandler(cfg.StorageHandler))
	}

	return e
}
