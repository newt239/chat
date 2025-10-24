package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	authuc "github.com/example/chat/internal/usecase/auth"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	userIDKey           = "userID"
	userEmailKey        = "userEmail"
)

func AuthMiddleware(jwtService authuc.JWTService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(authorizationHeader)
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authorization header required"})
			}

			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
			}

			token := strings.TrimPrefix(authHeader, bearerPrefix)
			claims, err := jwtService.VerifyToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			}

			c.Set(userIDKey, claims.UserID)
			c.Set(userEmailKey, claims.Email)
			return next(c)
		}
	}
}

func GetUserID(c echo.Context) (string, bool) {
	userID := c.Get(userIDKey)
	if userID == nil {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

func GetUserEmail(c echo.Context) (string, bool) {
	email := c.Get(userEmailKey)
	if email == nil {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}
