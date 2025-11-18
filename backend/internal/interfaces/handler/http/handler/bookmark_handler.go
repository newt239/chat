package handler

import (
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/labstack/echo/v4"
	"github.com/newt239/chat/internal/usecase/bookmark"
)

type BookmarkHandler struct {
	bookmarkUC bookmark.BookmarkUseCase
}

func NewBookmarkHandler(bookmarkUC bookmark.BookmarkUseCase) *BookmarkHandler {
	return &BookmarkHandler{
		bookmarkUC: bookmarkUC,
	}
}

func (h *BookmarkHandler) ListBookmarks(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	bookmarks, err := h.bookmarkUC.ListBookmarks(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "ブックマーク一覧の取得に失敗しました: " + err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"bookmarks": bookmarks.Bookmarks})
}

// AddBookmark はメッセージにブックマークを追加します (ServerInterface用)
func (h *BookmarkHandler) AddBookmark(c echo.Context, messageId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	input := bookmark.AddBookmarkInput{
		UserID:    userID,
		MessageID: messageId.String(),
	}

	err := h.bookmarkUC.AddBookmark(c.Request().Context(), input)
	if err != nil {
		switch err {
		case bookmark.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case bookmark.ErrBookmarkExists:
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case bookmark.ErrUnauthorized:
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "ブックマークの追加に失敗しました: " + err.Error()})
		}
	}

	return c.NoContent(http.StatusCreated)
}

// RemoveBookmark はメッセージからブックマークを削除します (ServerInterface用)
func (h *BookmarkHandler) RemoveBookmark(c echo.Context, messageId openapi_types.UUID) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "ユーザーが認証されていません"})
	}

	input := bookmark.RemoveBookmarkInput{
		UserID:    userID,
		MessageID: messageId.String(),
	}

	err := h.bookmarkUC.RemoveBookmark(c.Request().Context(), input)
	if err != nil {
		switch err {
		case bookmark.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case bookmark.ErrUnauthorized:
			return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
		default:
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "ブックマークの削除に失敗しました: " + err.Error()})
		}
	}

	return c.NoContent(http.StatusOK)
}
