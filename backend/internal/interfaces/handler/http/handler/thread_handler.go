package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/newt239/chat/internal/infrastructure/utils"
	"github.com/newt239/chat/internal/openapi_gen"
	threaduc "github.com/newt239/chat/internal/usecase/thread"
)

type ThreadHandler struct {
	ThreadLister *threaduc.ThreadLister
	ThreadReader *threaduc.ThreadReader
}

func (h *ThreadHandler) MarkThreadRead(ctx echo.Context, threadId openapi_types.UUID) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	input := threaduc.MarkThreadReadInput{
		UserID:   userID,
		ThreadID: threadId.String(),
	}

	err := h.ThreadReader.MarkThreadRead(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *ThreadHandler) GetParticipatingThreads(ctx echo.Context, workspaceId string, params openapi.GetParticipatingThreadsParams) error {
	userID, ok := ctx.Get("userID").(string)
	if !ok {
		return utils.HandleAuthError()
	}

	var cursorLastActivityAt *time.Time
	var cursorThreadID *string

	if params.CursorLastActivityAt != nil {
		cursorLastActivityAt = params.CursorLastActivityAt
	}

	if params.CursorThreadId != nil {
		cursorThreadIDStr := params.CursorThreadId.String()
		cursorThreadID = &cursorThreadIDStr
	}

	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}

	input := threaduc.ListParticipatingThreadsInput{
		WorkspaceID:          workspaceId,
		UserID:               userID,
		CursorLastActivityAt: cursorLastActivityAt,
		CursorThreadID:       cursorThreadID,
		Limit:                limit,
	}

	output, err := h.ThreadLister.ListParticipatingThreads(ctx.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return ctx.JSON(http.StatusOK, output)
}
