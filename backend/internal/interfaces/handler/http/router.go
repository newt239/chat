package http

import (
	"net"
	"net/http"
	"strings"

	connectcors "connectrpc.com/cors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
)

type RouterConfig struct {
	AllowedOrigins []string
	// X-Forwarded-For を信頼するプロキシの CIDR。空ならループバックとプライベートネットワークを信頼する
	TrustedProxies []string

	WebSocketHub   *websocket.Hub
	TicketConsumer websocket.TicketConsumer
	RPCHandler     http.Handler
	WebhookPoster  WebhookPoster
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
	e.IPExtractor = echo.ExtractIPFromXFFHeader(trustOptions(cfg.TrustedProxies)...)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     append(connectcors.AllowedMethods(), http.MethodPut),
		AllowHeaders:     append(connectcors.AllowedHeaders(), echo.HeaderAuthorization, "X-Chat-Client"),
		ExposeHeaders:    connectcors.ExposedHeaders(),
		AllowCredentials: true,
	}))
	// 開発用のローカルストレージはファイル本体を受け取るため上限をかけない
	e.Use(middleware.BodyLimitWithConfig(middleware.BodyLimitConfig{
		Limit:   "1M",
		Skipper: func(c echo.Context) bool { return strings.HasPrefix(c.Request().URL.Path, "/storage/") },
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

	e.GET("/ws", websocket.Handler(cfg.WebSocketHub, cfg.TicketConsumer, cfg.AllowedOrigins))

	e.Any("/chat.v1.*", echo.WrapHandler(cfg.RPCHandler), withRealIP)

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

// trustOptions は X-Forwarded-For をたどるときに信頼するプロキシを決めます
func trustOptions(trustedProxies []string) []echo.TrustOption {
	if len(trustedProxies) == 0 {
		return nil
	}
	opts := []echo.TrustOption{echo.TrustLoopback(false), echo.TrustLinkLocal(false), echo.TrustPrivateNet(false)}
	for _, cidr := range trustedProxies {
		if _, ipNet, err := net.ParseCIDR(cidr); err == nil {
			opts = append(opts, echo.TrustIPRange(ipNet))
		}
	}
	return opts
}

// withRealIP は Connect のハンドラが接続元として読む RemoteAddr を、信頼するプロキシを考慮した IP に置き換えます
func withRealIP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().RemoteAddr = net.JoinHostPort(c.RealIP(), "0")
		return next(c)
	}
}
