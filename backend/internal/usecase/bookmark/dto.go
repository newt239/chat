package bookmark

import "time"

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

type ListBookmarksOutput struct {
	Bookmarks []BookmarkOutput
}
