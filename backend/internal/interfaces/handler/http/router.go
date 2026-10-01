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
	// 未設定ならプロセス内で数える
	WebhookRateLimiter RateLimiter
	// false を返す間は readiness probe に 503 を返し、停止前に新しい接続を受けないようにする
	Ready func() bool
	// 未設定ならネイティブアプリの Google ログインの経路を作らない
	GoogleOAuth GoogleOAuthFlow
	// 開発用のローカルストレージを使うときだけ設定する
	StorageHandler http.Handler
	// アイコン画像の配信に使う
	Storage       service.StorageService
	StorageConfig service.StorageConfig
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
	e.GET("/readyz", func(c echo.Context) error {
		if cfg.Ready != nil && !cfg.Ready() {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "shutting_down"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ws", websocket.Handler(cfg.WebSocketHub, cfg.JWTService, cfg.WorkspaceRepository, cfg.ChannelAccess, cfg.AllowedOrigins))

	e.Any("/chat.v1.*", echo.WrapHandler(cfg.RPCHandler))

	limiter := cfg.WebhookRateLimiter
	if limiter == nil {
		limiter = newRateLimiter(WebhookRatePerSecond, WebhookBurst)
	}
	e.POST("/webhooks/:id/:token", webhookHandler(cfg.WebhookPoster, limiter))

	if cfg.GoogleOAuth != nil {
		e.GET("/oauth/google/start", googleOAuthStartHandler(cfg.GoogleOAuth))
		e.GET("/oauth/google/callback", googleOAuthCallbackHandler(cfg.GoogleOAuth))
	}

	e.GET("/images/*", imageHandler(cfg.Storage, cfg.StorageConfig))

	if cfg.StorageHandler != nil {
		e.Any("/storage/*", echo.WrapHandler(cfg.StorageHandler))
	}

	return e
}
