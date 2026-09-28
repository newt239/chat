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
	Thumbnail  *ThumbnailInput
	ExpiresMin int
}

// ThumbnailInput は動画と一緒にアップロードするサムネイル画像です
type ThumbnailInput struct {
	MimeType  string
	SizeBytes int64
	Width     int32
	Height    int32
}

type PresignOutput struct {
	AttachmentID string
	UploadURL    string
	// サムネイルを指定したときだけ返します
	ThumbnailUploadURL *string
	StorageKey         string
	ExpiresAt          time.Time
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
