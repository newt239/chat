package attachment

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type PresignInput struct {
	UserID     string
	ChannelID  string
	FileName   string
	MimeType   string
	SizeBytes  int64
	Media      entity.MediaMetadata
	ExpiresMin int
}

type PresignOutput struct {
	AttachmentID string
	UploadURL    string
	StorageKey   string
	ExpiresAt    time.Time
}

type AttachmentOutput struct {
	ID         string
	MessageID  *string
	UploaderID string
	ChannelID  string
	FileName   string
	MimeType   string
	SizeBytes  int64
	Media      entity.MediaMetadata
	Status     string
	CreatedAt  time.Time
}

type DownloadURLOutput struct {
	URL       string
	ExpiresIn int
}
