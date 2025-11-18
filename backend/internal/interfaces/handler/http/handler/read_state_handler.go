package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

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

// UpdateReadState implements ServerInterface.UpdateReadState
func (h *ReadStateHandler) UpdateReadState(ctx echo.Context, channelId openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req UpdateReadStateRequest
	if err := ctx.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	lastReadAt, parseErr := time.Parse(time.RFC3339, req.LastReadAt)
	if parseErr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "last_read_atの形式が不正です")
	}

	input := readstateuc.UpdateReadStateInput{
		ChannelID:  channelId.String(),
		UserID:     userID,
		LastReadAt: lastReadAt,
	}

	err := h.readStateUC.UpdateReadState(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// GetUnreadCount implements ServerInterface.GetUnreadCount
func (h *ReadStateHandler) GetUnreadCount(ctx echo.Context, channelId openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := readstateuc.GetUnreadCountInput{
		ChannelID: channelId.String(),
		UserID:    userID,
	}

	count, err := h.readStateUC.GetUnreadCount(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, count)
}
