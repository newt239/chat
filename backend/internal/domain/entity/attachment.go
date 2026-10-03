package entity

import "time"

type AttachmentStatus string

const (
	AttachmentStatusPending  AttachmentStatus = "pending"
	AttachmentStatusAttached AttachmentStatus = "attached"
)

type Attachment struct {
	ID         string
	MessageID  *string
	UploaderID string
	ChannelID  string
	FileName   string
	MimeType   string
	SizeBytes  int64
	Media      MediaMetadata
	StorageKey string
	Status     AttachmentStatus
	CreatedAt  time.Time
}

// MediaMetadata は画像・動画・音声の表示に使う寸法と再生時間です
type MediaMetadata struct {
	Width           *int32
	Height          *int32
	DurationSeconds *float64
	Thumbnail       *Thumbnail
}

// Thumbnail は動画の再生前に表示する画像です。アップロード時にクライアントが切り出します
type Thumbnail struct {
	StorageKey string
	Width      int32
	Height     int32
}
