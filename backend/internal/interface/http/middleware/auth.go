package middleware

import (
	"net/http"
	"strings"

	"github.com/example/chat/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
)

const (
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
	userIDKey           = "userID"
	userEmailKey        = "userEmail"
)

func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(authorizationHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, bearerPrefix)
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Set(userEmailKey, claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(userIDKey)
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(userEmailKey)
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}
