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
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	limitParam := c.QueryParam("limit")
	limit := 0
	if limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit parameter"})
		}
		limit = parsed
	}

	var sincePtr *time.Time
	if sinceParam := c.QueryParam("since"); sinceParam != "" {
		parsed, err := time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid since parameter"})
		}
		sincePtr = &parsed
	}

	var untilPtr *time.Time
	if untilParam := c.QueryParam("until"); untilParam != "" {
		parsed, err := time.Parse(time.RFC3339, untilParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid until parameter"})
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
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list messages"})
		}
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
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	var req CreateMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	output, err := h.messageUseCase.CreateMessage(c.Request().Context(), message.CreateMessageInput{
		ChannelID:     channelID,
		UserID:        userID,
		Body:          req.Body,
		ParentID:      req.ParentID,
		AttachmentIDs: req.AttachmentIDs,
	})
	if err != nil {
		switch err {
		case message.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case message.ErrParentMessageNotFound:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create message"})
		}
	}

	return c.JSON(http.StatusCreated, output)
}

// GetThreadReplies godoc
// @Summary Get thread replies
// @Description Returns all replies for a specific thread
// @Tags message,thread
// @Produce json
// @Param messageId path string true "Parent Message ID"
// @Success 200 {object} message.GetThreadRepliesOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId}/thread [get]
// @Security BearerAuth
func (h *MessageHandler) GetThreadReplies(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
	}

	output, err := h.messageUseCase.GetThreadReplies(c.Request().Context(), message.GetThreadRepliesInput{
		MessageID: messageID,
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case message.ErrParentMessageNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get thread replies"})
		}
	}

	return c.JSON(http.StatusOK, output)
}

// GetThreadMetadata godoc
// @Summary Get thread metadata
// @Description Returns metadata for a specific thread (reply count, participants, etc.)
// @Tags message,thread
// @Produce json
// @Param messageId path string true "Parent Message ID"
// @Success 200 {object} message.ThreadMetadataOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId}/thread/metadata [get]
// @Security BearerAuth
func (h *MessageHandler) GetThreadMetadata(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
	}

	output, err := h.messageUseCase.GetThreadMetadata(c.Request().Context(), message.GetThreadMetadataInput{
		MessageID: messageID,
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case message.ErrParentMessageNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get thread metadata"})
		}
	}

	return c.JSON(http.StatusOK, output)
}

// ListMessagesWithThread godoc
// @Summary List channel messages with thread metadata
// @Description Returns messages in the specified channel with thread metadata
// @Tags message,thread
// @Produce json
// @Param channelId path string true "Channel ID"
// @Param limit query int false "Number of messages to return (1-100, default 50)"
// @Param since query string false "Return messages created after this timestamp (RFC3339)"
// @Param until query string false "Return messages created before this timestamp (RFC3339)"
// @Success 200 {array} message.MessageWithThreadOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/messages/with-threads [get]
// @Security BearerAuth
func (h *MessageHandler) ListMessagesWithThread(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
	}

	limitParam := c.QueryParam("limit")
	limit := 0
	if limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit parameter"})
		}
		limit = parsed
	}

	var sincePtr *time.Time
	if sinceParam := c.QueryParam("since"); sinceParam != "" {
		parsed, err := time.Parse(time.RFC3339, sinceParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid since parameter"})
		}
		sincePtr = &parsed
	}

	var untilPtr *time.Time
	if untilParam := c.QueryParam("until"); untilParam != "" {
		parsed, err := time.Parse(time.RFC3339, untilParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid until parameter"})
		}
		untilPtr = &parsed
	}

	output, err := h.messageUseCase.ListMessagesWithThread(c.Request().Context(), message.ListMessagesInput{
		ChannelID: channelID,
		UserID:    userID,
		Limit:     limit,
		Since:     sincePtr,
		Until:     untilPtr,
	})
	if err != nil {
		switch err {
		case message.ErrChannelNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list messages with threads"})
		}
	}

	return c.JSON(http.StatusOK, output)
}

// UpdateMessageRequest はメッセージ更新リクエストを表します
type UpdateMessageRequest struct {
	Body string `json:"body" validate:"required,min=1,max=10000"`
}

// UpdateMessage godoc
// @Summary Update a message
// @Description Updates a message body (author or admin only)
// @Tags message
// @Accept json
// @Produce json
// @Param messageId path string true "Message ID"
// @Param body body UpdateMessageRequest true "Message body"
// @Success 200 {object} message.MessageOutput
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId} [patch]
// @Security BearerAuth
func (h *MessageHandler) UpdateMessage(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
	}

	var req UpdateMessageRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if req.Body == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message body is required"})
	}

	if len(req.Body) > 10000 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message body is too long"})
	}

	output, err := h.messageUseCase.UpdateMessage(c.Request().Context(), message.UpdateMessageInput{
		MessageID: messageID,
		EditorID:  userID,
		Body:      req.Body,
	})
	if err != nil {
		switch err {
		case message.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case message.ErrCannotEditDeleted:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update message"})
		}
	}

	return c.JSON(http.StatusOK, output)
}

// DeleteMessage godoc
// @Summary Delete a message
// @Description Deletes a message (author or admin only)
// @Tags message
// @Param messageId path string true "Message ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId} [delete]
// @Security BearerAuth
func (h *MessageHandler) DeleteMessage(c echo.Context) error {
	userID, err := requireUserID(c)
	if err != nil {
		return err
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
	}

	err = h.messageUseCase.DeleteMessage(c.Request().Context(), message.DeleteMessageInput{
		MessageID:  messageID,
		ExecutorID: userID,
	})
	if err != nil {
		switch err {
		case message.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case message.ErrUnauthorized:
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case message.ErrMessageAlreadyDeleted:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete message"})
		}
	}

	return c.NoContent(http.StatusNoContent)
}
