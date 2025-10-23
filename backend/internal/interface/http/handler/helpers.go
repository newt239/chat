package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/interface/http/middleware"
)

func requireUserID(c *gin.Context) (string, bool) {
	userID, exists := middleware.GetUserID(c)
	if !exists || userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return "", false
	}
	return userID, true
}
