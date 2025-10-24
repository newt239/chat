package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/example/chat/internal/usecase/message"
)

type MessageHandler struct {
	messageUseCase message.MessageUseCase
}

func NewMessageHandler(messageUseCase message.MessageUseCase) *MessageHandler {
	return &MessageHandler{messageUseCase: messageUseCase}
}

// ListMessages godoc
// @Summary List channel messages
// @Description Returns messages in the specified channel
// @Tags message
// @Produce json
// @Param channelId path string true "Channel ID"
// @Param limit query int false "Number of messages to return err (1-100, default 50)"
// @Param since query string false "Return messages created after this timestamp (RFC3339)"
// @Param until query string false "Return messages created before this timestamp (RFC3339)"
// @Success 200 {object} message.ListMessagesOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/messages [get]
// @Security BearerAuth
func (h *MessageHandler) ListMessages(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
		return err
	}

	limitParam := c.QueryParam("limit")
	limit := 0
	if limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit parameter"})
			return err
		}
		limit = parsed
	}

	var sincePtr *time.Time
	if sinceParam := c.QueryParam("since"); sinceParam != "" {
		parsed, err := time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid since parameter"})
			return err
		}
		sincePtr = &parsed
	}

	var untilPtr *time.Time
	if untilParam := c.QueryParam("until"); untilParam != "" {
		parsed, err := time.Parse(time.RFC3339, untilParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid until parameter"})
			return err
		}
		untilPtr = &parsed
	}

	output, err := h.messageUseCase.ListMessages(c.Request().Context(), message.ListMessagesInput{
		ChannelID: channelID,
		UserID:    userID,
		Limit:     limit,
		Since:     sincePtr,
		Until:     untilPtr,
	})
	if err != nil {
		switch err {
		case message.ErrChannelNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list messages"})
		}
		return err
	}

	return c.JSON(http.StatusOK, output)
}

// CreateMessage godoc
// @Summary Create a message
// @Description Creates a new message within the specified channel
// @Tags message
// @Accept json
// @Produce json
// @Param channelId path string true "Channel ID"
// @Param request body CreateMessageRequest true "Create message request"
// @Success 201 {object} message.MessageOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/messages [post]
// @Security BearerAuth
func (h *MessageHandler) CreateMessage(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
		return err
	}

	var req CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	output, err := h.messageUseCase.CreateMessage(c.Request().Context(), message.CreateMessageInput{
		ChannelID: channelID,
		UserID:    userID,
		Body:      req.Body,
		ParentID:  req.ParentID,
	})
	if err != nil {
		switch err {
		case message.ErrChannelNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case message.ErrParentMessageNotFound:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create message"})
		}
		return err
	}

	return c.JSON(http.StatusCreated, output)
}
