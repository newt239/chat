package customemoji

import messageuc "github.com/newt239/chat/internal/usecase/message"

type ListInput struct {
	WorkspaceID string
	UserID      string
}

type PresignInput struct {
	WorkspaceID string
	UserID      string
	ContentType string
}

type CreateInput struct {
	WorkspaceID string
	UserID      string
	Name        string
	UploadID    string
}

type DeleteInput struct {
	WorkspaceID string
	UserID      string
	EmojiID     string
}

type Output struct {
	ID        string
	Name      string
	ImageURL  string
	CreatedBy messageuc.UserInfo
	CanDelete bool
}

type PresignOutput struct {
	UploadID  string
	UploadURL string
}
