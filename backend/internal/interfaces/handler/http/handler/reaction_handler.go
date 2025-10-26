package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

type ReactionHandler struct {
	reactionUC reactionuc.ReactionUseCase
}

func NewReactionHandler(reactionUC reactionuc.ReactionUseCase) *ReactionHandler {
	return &ReactionHandler{reactionUC: reactionUC}
}

// AddReactionRequest はリアクション追加リクエストの構造体です
type AddReactionRequest struct {
	Emoji string `json:"emoji" validate:"required,min=1"`
}

// ListReactions はリアクション一覧を取得します
func (h *ReactionHandler) ListReactions(c echo.Context) error {
	messageID := c.Param("messageId")
	if messageID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Message ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	reactions, err := h.reactionUC.ListReactions(c.Request().Context(), messageID, userID)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, reactions)
}

// AddReaction はリアクションを追加します
func (h *ReactionHandler) AddReaction(c echo.Context) error {
	messageID := c.Param("messageId")
	if messageID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Message ID is required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	var req AddReactionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := reactionuc.AddReactionInput{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     req.Emoji,
	}

	err := h.reactionUC.AddReaction(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusCreated)
}

// RemoveReaction はリアクションを削除します
func (h *ReactionHandler) RemoveReaction(c echo.Context) error {
	messageID := c.Param("messageId")
	emoji := c.Param("emoji")
	if messageID == "" || emoji == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Message ID and emoji are required")
	}

	userID, ok := c.Get("userID").(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	input := reactionuc.RemoveReactionInput{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	err := h.reactionUC.RemoveReaction(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
