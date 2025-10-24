package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/interface/http/middleware"
)

func requireUserID(c echo.Context) (string, error) {
	userID, exists := middleware.GetUserID(c)
	if !exists || userID == "" {
		return "", c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
	}
	return userID, nil
}
