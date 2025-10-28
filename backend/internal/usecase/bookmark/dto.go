package bookmark

import (
	"time"

	"github.com/newt239/chat/internal/usecase/message"
)

type AddBookmarkInput struct {
	UserID    string
	MessageID string
}

type RemoveBookmarkInput struct {
	UserID    string
	MessageID string
}

type BookmarkOutput struct {
	UserID    string
	MessageID string
	CreatedAt time.Time
}

type BookmarkWithMessageOutput struct {
	UserID    string                `json:"userId"`
	Message   message.MessageOutput `json:"message"`
	CreatedAt time.Time             `json:"createdAt"`
}

type ListBookmarksOutput struct {
	Bookmarks []BookmarkWithMessageOutput `json:"bookmarks"`
}
