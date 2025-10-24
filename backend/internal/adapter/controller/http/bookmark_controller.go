package http

import (
	"github.com/example/chat/internal/interface/http/handler"
	"github.com/labstack/echo/v4"
)

type BookmarkController struct {
	handler *handler.BookmarkHandler
}

func NewBookmarkController(handler *handler.BookmarkHandler) *BookmarkController {
	return &BookmarkController{
		handler: handler,
	}
}

func (c *BookmarkController) RegisterRoutes(router *echo.Group) {
	bookmarks := router.Group("/bookmarks")
	bookmarks.GET("", c.handler.ListBookmarks)

	messages := router.Group("/messages")
	messages.POST("/:messageId/bookmarks", c.handler.AddBookmark)
	messages.DELETE("/:messageId/bookmarks", c.handler.RemoveBookmark)
}
