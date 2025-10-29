package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/usecase/pin"
)

type PinHandler struct {
	uc pin.PinUseCase
}

func NewPinHandler(uc pin.PinUseCase) *PinHandler { return &PinHandler{uc: uc} }

type PinRequest struct {
	MessageID string `json:"messageId" validate:"required,uuid4"`
}

// POST /channels/:channelId/pins
func (h *PinHandler) CreatePin(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	channelID := c.Param("channelId")
	var req PinRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	input := pin.PinMessageInput{ChannelID: channelID, MessageID: req.MessageID, UserID: userID}
	if err := h.uc.PinMessage(c.Request().Context(), input); err != nil {
		switch err {
		case pin.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		case pin.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		default:
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
	}
	return c.NoContent(http.StatusOK)
}

// DELETE /channels/:channelId/pins/:messageId
func (h *PinHandler) DeletePin(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	channelID := c.Param("channelId")
	messageID := c.Param("messageId")
	input := pin.UnpinMessageInput{ChannelID: channelID, MessageID: messageID, UserID: userID}
	if err := h.uc.UnpinMessage(c.Request().Context(), input); err != nil {
		switch err {
		case pin.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		case pin.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	return c.NoContent(http.StatusOK)
}

// GET /channels/:channelId/pins?limit=100&cursor=...
func (h *PinHandler) ListPins(c echo.Context) error {
	userID, _ := c.Get("userID").(string)
	if userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	channelID := c.Param("channelId")
	limit := 100
	if l := c.QueryParam("limit"); l != "" {
		if v, err := atoiSafe(l); err == nil {
			limit = v
		}
	}
	cursor := c.QueryParam("cursor")
	var curPtr *string
	if cursor != "" {
		curPtr = &cursor
	}

	input := pin.ListPinsInput{ChannelID: channelID, UserID: userID, Limit: limit, Cursor: curPtr}
	out, err := h.uc.ListPins(c.Request().Context(), input)
	if err != nil {
		if err == pin.ErrUnauthorized {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "forbidden"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"pins": out.Pins, "nextCursor": out.NextCursor})
}

func atoiSafe(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
