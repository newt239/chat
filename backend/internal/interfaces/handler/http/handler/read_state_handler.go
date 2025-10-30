package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	readstateuc "github.com/newt239/chat/internal/usecase/readstate"
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
		return echo.NewHTTPError(http.StatusBadRequest, "チャンネルIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
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
		return echo.NewHTTPError(http.StatusBadRequest, "チャンネルIDは必須です")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req UpdateReadStateRequest
	if err := c.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	lastReadAt, parseErr := time.Parse(time.RFC3339, req.LastReadAt)
	if parseErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "last_read_atの形式が不正です")
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
