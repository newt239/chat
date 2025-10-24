package handler

import (
	"net/http"
	"time"

	readstateuc "github.com/example/chat/internal/usecase/readstate"
	"github.com/labstack/echo/v4"
)

type ReadStateHandler struct {
	readStateUC readstateuc.ReadStateUseCase
}

func NewReadStateHandler(readStateUC readstateuc.ReadStateUseCase) *ReadStateHandler {
	return &ReadStateHandler{readStateUC: readStateUC}
}

// UpdateReadStateRequest は既読状態更新リクエストの構造体です
type UpdateReadStateRequest struct {
	LastReadAt string `json:"last_read_at" validate:"required"`
}

// GetUnreadCount は未読メッセージ数を取得します
func (h *ReadStateHandler) GetUnreadCount(c echo.Context) error {
	channelID := c.Param("channelId")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := readstateuc.GetUnreadCountInput{
		ChannelID: channelID,
		UserID:    userID,
	}

	count, err := h.readStateUC.GetUnreadCount(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, count)
}

// UpdateReadState は既読状態を更新します
func (h *ReadStateHandler) UpdateReadState(c echo.Context) error {
	channelID := c.Param("channelId")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req UpdateReadStateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	lastReadAt, parseErr := time.Parse(time.RFC3339, req.LastReadAt)
	if parseErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid last_read_at format")
	}

	input := readstateuc.UpdateReadStateInput{
		ChannelID:  channelID,
		UserID:     userID,
		LastReadAt: lastReadAt,
	}

	err := h.readStateUC.UpdateReadState(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusOK)
}
