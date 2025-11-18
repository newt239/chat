package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	openapi "github.com/newt239/chat/internal/openapi_gen"
	"github.com/newt239/chat/internal/usecase/pin"
)

type PinHandler struct {
	UC pin.PinUseCase
}

type PinRequest struct {
	MessageID string `json:"messageId" validate:"required,uuid4"`
}

func (h *PinHandler) ListPins(ctx echo.Context, channelId openapi_types.UUID, params openapi.ListPinsParams) error {
	userID, _ := ctx.Get("userID").(string)
	if userID == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	limit := 100
	if params.Limit != nil {
		limit = *params.Limit
	}
	var curPtr *string
	if params.Cursor != nil {
		curPtr = params.Cursor
	}

	input := pin.ListPinsInput{ChannelID: channelId.String(), UserID: userID, Limit: limit, Cursor: curPtr}
	out, err := h.UC.ListPins(ctx.Request().Context(), input)
	if err != nil {
		if err == pin.ErrUnauthorized {
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, map[string]interface{}{"pins": out.Pins, "nextCursor": out.NextCursor})
}

func (h *PinHandler) CreatePin(ctx echo.Context, channelId openapi_types.UUID) error {
	userID, _ := ctx.Get("userID").(string)
	if userID == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	var req PinRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := ctx.Validate(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	input := pin.PinMessageInput{ChannelID: channelId.String(), MessageID: req.MessageID, UserID: userID}
	if err := h.UC.PinMessage(ctx.Request().Context(), input); err != nil {
		switch err {
		case pin.ErrUnauthorized:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		case pin.ErrMessageNotFound:
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		default:
			return ctx.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
	}
	return ctx.NoContent(http.StatusOK)
}

func (h *PinHandler) DeletePin(ctx echo.Context, channelId openapi_types.UUID, messageId openapi_types.UUID) error {
	userID, _ := ctx.Get("userID").(string)
	if userID == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	input := pin.UnpinMessageInput{ChannelID: channelId.String(), MessageID: messageId.String(), UserID: userID}
	if err := h.UC.UnpinMessage(ctx.Request().Context(), input); err != nil {
		switch err {
		case pin.ErrUnauthorized:
			return ctx.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		case pin.ErrMessageNotFound:
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		default:
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	return ctx.NoContent(http.StatusOK)
}
