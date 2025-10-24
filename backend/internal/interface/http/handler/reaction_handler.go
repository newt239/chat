package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/example/chat/internal/usecase/reaction"
)

type ReactionHandler struct {
	reactionUseCase reaction.ReactionUseCase
}

func NewReactionHandler(reactionUseCase reaction.ReactionUseCase) *ReactionHandler {
	return &ReactionHandler{reactionUseCase: reactionUseCase}
}

// ListReactions godoc
// @Summary List reactions for a message
// @Description Returns all reactions for the specified message
// @Tags reaction
// @Produce json
// @Param messageId path string true "Message ID"
// @Success 200 {object} reaction.ListReactionsOutput
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId}/reactions [get]
// @Security BearerAuth
func (h *ReactionHandler) ListReactions(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
		return
	}

	output, err := h.reactionUseCase.ListReactions(c.Request.Context(), messageID, userID)
	if err != nil {
		switch err {
		case reaction.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case reaction.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list reactions"})
		}
		return
	}

	c.JSON(http.StatusOK, output)
}

// AddReaction godoc
// @Summary Add a reaction to a message
// @Description Adds a reaction (emoji) to the specified message
// @Tags reaction
// @Accept json
// @Produce json
// @Param messageId path string true "Message ID"
// @Param request body AddReactionRequest true "Add reaction request"
// @Success 201
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId}/reactions [post]
// @Security BearerAuth
func (h *ReactionHandler) AddReaction(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
		return
	}

	var req AddReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.reactionUseCase.AddReaction(c.Request.Context(), reaction.AddReactionInput{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     req.Emoji,
	})
	if err != nil {
		switch err {
		case reaction.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case reaction.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to add reaction"})
		}
		return
	}

	c.Status(http.StatusCreated)
}

// RemoveReaction godoc
// @Summary Remove a reaction from a message
// @Description Removes the user's reaction (emoji) from the specified message
// @Tags reaction
// @Produce json
// @Param messageId path string true "Message ID"
// @Param emoji path string true "Emoji"
// @Success 200
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/messages/{messageId}/reactions/{emoji} [delete]
// @Security BearerAuth
func (h *ReactionHandler) RemoveReaction(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Message ID is required"})
		return
	}

	emoji := c.Param("emoji")
	if emoji == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Emoji is required"})
		return
	}

	err := h.reactionUseCase.RemoveReaction(c.Request.Context(), reaction.RemoveReactionInput{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	})
	if err != nil {
		switch err {
		case reaction.ErrMessageNotFound:
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case reaction.ErrUnauthorized:
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to remove reaction"})
		}
		return
	}

	c.Status(http.StatusOK)
}
