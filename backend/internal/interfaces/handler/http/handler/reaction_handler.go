package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/infrastructure/utils"
	"github.com/newt239/chat/internal/openapi_gen"
	reactionuc "github.com/newt239/chat/internal/usecase/reaction"
)

type ReactionHandler struct {
	ReactionUC reactionuc.ReactionUseCase
}

// ListReactions implements ServerInterface.ListReactions
func (h *ReactionHandler) ListReactions(ctx echo.Context, messageId openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	reactions, err := h.ReactionUC.ListReactions(ctx.Request().Context(), messageId.String(), userID)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, reactions)
}

// AddReaction implements ServerInterface.AddReaction
func (h *ReactionHandler) AddReaction(ctx echo.Context, messageId openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var req openapi.AddReactionRequest
	if err := ctx.Bind(&req); err != nil {
		return utils.HandleBindError(err)
	}

	if err := ctx.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	input := reactionuc.AddReactionInput{
		MessageID: messageId.String(),
		UserID:    userID,
		Emoji:     req.Emoji,
	}

	err := h.ReactionUC.AddReaction(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

// RemoveReaction implements ServerInterface.RemoveReaction
func (h *ReactionHandler) RemoveReaction(ctx echo.Context, messageId openapi_types.UUID, emoji string) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := reactionuc.RemoveReactionInput{
		MessageID: messageId.String(),
		UserID:    userID,
		Emoji:     emoji,
	}

	err := h.ReactionUC.RemoveReaction(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}
