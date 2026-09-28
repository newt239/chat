package http

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

type RouterConfig struct {
	JWTService     authuc.JWTService
	AllowedOrigins []string

	WebSocketHub        *websocket.Hub
	WorkspaceRepository repository.WorkspaceRepository
	RPCHandler          http.Handler
}

func NewRouter(cfg RouterConfig) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     connectcors.AllowedMethods(),
		AllowHeaders:     append(connectcors.AllowedHeaders(), echo.HeaderAuthorization),
		ExposeHeaders:    connectcors.ExposedHeaders(),
		AllowCredentials: true,
	}))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/ws", websocket.Handler(cfg.WebSocketHub, cfg.JWTService, cfg.WorkspaceRepository, cfg.AllowedOrigins))

	e.Any("/chat.v1.*", echo.WrapHandler(cfg.RPCHandler))

	return e
}
