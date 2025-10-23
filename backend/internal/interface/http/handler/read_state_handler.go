package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/usecase/readstate"
)

type ReadStateHandler struct {
	readStateUseCase readstate.ReadStateUseCase
}

func NewReadStateHandler(readStateUseCase readstate.ReadStateUseCase) *ReadStateHandler {
	return &ReadStateHandler{readStateUseCase: readStateUseCase}
}

// GetUnreadCount godoc
// @Summary Get unread message count for a channel
// @Tags message
// @Produce json
// @Param channelId path string true "Channel ID"
// @Success 200 {object} readstate.UnreadCountOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/unread_count [get]
// @Security BearerAuth
func (h *ReadStateHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
		return
	}

	output, err := h.readStateUseCase.GetUnreadCount(readstate.GetUnreadCountInput{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		switch err {
		case readstate.ErrChannelNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case readstate.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get unread count"})
		}
		return
	}

	c.JSON(http.StatusOK, output)
}

// UpdateReadState godoc
// @Summary Update channel read state
// @Tags message
// @Accept json
// @Param channelId path string true "Channel ID"
// @Param request body UpdateReadStateRequest true "Update read state request"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/channels/{channelId}/reads [post]
// @Security BearerAuth
func (h *ReadStateHandler) UpdateReadState(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	channelID := c.Param("channelId")
	if channelID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Channel ID is required"})
		return
	}

	var req UpdateReadStateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	parsedTime, err := time.Parse(time.RFC3339, req.LastReadAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid lastReadAt format"})
		return
	}

	if err := h.readStateUseCase.UpdateReadState(readstate.UpdateReadStateInput{
		ChannelID:  channelID,
		UserID:     userID,
		LastReadAt: parsedTime,
	}); err != nil {
		switch err {
		case readstate.ErrChannelNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case readstate.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update read state"})
		}
		return
	}

	c.Status(http.StatusOK)
}
