package handler

import (
	"net/http"

	"github.com/example/chat/internal/usecase/bookmark"
	"github.com/labstack/echo/v4"
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
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	bookmarks, err := h.bookmarkUC.ListBookmarks(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"bookmarks": bookmarks.Bookmarks})
}

func (h *BookmarkHandler) AddBookmark(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "message ID is required"})
	}

	input := bookmark.AddBookmarkInput{
		UserID:    userID,
		MessageID: messageID,
	}

	err := h.bookmarkUC.AddBookmark(c.Request().Context(), input)
	if err != nil {
		switch err {
		case bookmark.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		case bookmark.ErrBookmarkExists:
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "bookmark already exists"})
		case bookmark.ErrUnauthorized:
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.NoContent(http.StatusCreated)
}

func (h *BookmarkHandler) RemoveBookmark(c echo.Context) error {
	userID, ok := c.Get("userID").(string)
	if !ok || userID == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	messageID := c.Param("messageId")
	if messageID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "message ID is required"})
	}

	input := bookmark.RemoveBookmarkInput{
		UserID:    userID,
		MessageID: messageID,
	}

	err := h.bookmarkUC.RemoveBookmark(c.Request().Context(), input)
	if err != nil {
		switch err {
		case bookmark.ErrMessageNotFound:
			return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
		case bookmark.ErrUnauthorized:
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		default:
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}

	return c.NoContent(http.StatusOK)
}
