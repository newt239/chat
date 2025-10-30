package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/infrastructure/utils"
	threaduc "github.com/newt239/chat/internal/usecase/thread"
)

type ThreadHandler struct {
	threadLister *threaduc.ThreadLister
	threadReader *threaduc.ThreadReader
}

func NewThreadHandler(
	threadLister *threaduc.ThreadLister,
	threadReader *threaduc.ThreadReader,
) *ThreadHandler {
	return &ThreadHandler{
		threadLister: threadLister,
		threadReader: threadReader,
	}
}

// GetParticipatingThreads は参加中スレッド一覧を取得します
func (h *ThreadHandler) GetParticipatingThreads(c echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	workspaceID := c.Param("workspaceId")
	if workspaceID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "workspaceIdは必須です")
	}

	// クエリパラメータの取得
	var cursorLastActivityAt *time.Time
	var cursorThreadID *string

	if cursorLastActivityAtStr := c.QueryParam("cursorLastActivityAt"); cursorLastActivityAtStr != "" {
		parsed, err := time.Parse(time.RFC3339, cursorLastActivityAtStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "cursorLastActivityAtの形式が不正です")
		}
		cursorLastActivityAt = &parsed
	}

	if cursorThreadIDStr := c.QueryParam("cursorThreadId"); cursorThreadIDStr != "" {
		cursorThreadID = &cursorThreadIDStr
	}

	limit := 20
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "limitの形式が不正です")
		}
		limit = parsedLimit
	}

	input := threaduc.ListParticipatingThreadsInput{
		WorkspaceID:          workspaceID,
		UserID:               userID,
		CursorLastActivityAt: cursorLastActivityAt,
		CursorThreadID:       cursorThreadID,
		Limit:                limit,
	}

	output, err := h.threadLister.ListParticipatingThreads(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.JSON(http.StatusOK, output)
}

// MarkThreadRead はスレッドを既読にします
func (h *ThreadHandler) MarkThreadRead(c echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return err
	}

	threadID := c.Param("threadId")
	if threadID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "threadIdは必須です")
	}

	input := threaduc.MarkThreadReadInput{
		UserID:   userID,
		ThreadID: threadID,
	}

	err = h.threadReader.MarkThreadRead(c.Request().Context(), input)
	if err != nil {
		return handleUseCaseError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
