package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type MessageHandler struct {
	messageUC messageuc.MessageUseCase
}

func NewMessageHandler(messageUC messageuc.MessageUseCase) *MessageHandler {
	return &MessageHandler{messageUC: messageUC}
}

// CreateMessageRequest はメッセージ作成リクエストの構造体です
type CreateMessageRequest struct {
	Body     string  `json:"body" validate:"required,min=1"`
	ParentID *string `json:"parentId,omitempty"`
}

// UpdateMessageRequest はメッセージ更新リクエストの構造体です
type UpdateMessageRequest struct {
	Body string `json:"body" validate:"required,min=1"`
}

// ListMessages はメッセージ一覧を取得します
func (h *MessageHandler) ListMessages(c echo.Context) error {
	channelID := c.Param("channelId")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	limitStr := c.QueryParam("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 日付フィルター
	sinceStr := c.QueryParam("since")
	since := ""
	if sinceStr != "" {
		since = sinceStr
	}

	untilStr := c.QueryParam("until")
	until := ""
	if untilStr != "" {
		until = untilStr
	}

	var sinceTime *time.Time
	if since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			sinceTime = &t
		}
	}

	var untilTime *time.Time
	if until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			untilTime = &t
		}
	}

	input := messageuc.ListMessagesInput{
		ChannelID: channelID,
		UserID:    userID,
		Limit:     limit,
		Since:     sinceTime,
		Until:     untilTime,
	}

	messages, err := h.messageUC.ListMessages(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, messages)
}

// CreateMessage はメッセージを作成します
func (h *MessageHandler) CreateMessage(c echo.Context) error {
	channelID := c.Param("channelId")
	if channelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Channel ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := messageuc.CreateMessageInput{
		ChannelID: channelID,
		UserID:    userID,
		Body:      req.Body,
		ParentID:  req.ParentID,
	}

	message, err := h.messageUC.CreateMessage(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusCreated, message)
}
