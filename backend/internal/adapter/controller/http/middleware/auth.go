package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	authuc "github.com/newt239/chat/internal/usecase/auth"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	userIDKey           = "userID"
	userEmailKey        = "userEmail"
)

// Auth は認証ミドルウェアを返します
func Auth(jwtService authuc.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(authorizationHeader)
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "authorization header required")
			}

			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			token := strings.TrimPrefix(authHeader, bearerPrefix)
			claims, err := jwtService.VerifyToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			c.Set(userIDKey, claims.UserID)
			c.Set(userEmailKey, claims.Email)

			return next(c)
		}
	}
}

// GetUserID はコンテキストからユーザーIDを取得します

// GetUserEmail はコンテキストからユーザーメールアドレスを取得します
