package entity

import "time"

type AttachmentStatus string

const (
	AttachmentStatusPending  AttachmentStatus = "pending"
	AttachmentStatusAttached AttachmentStatus = "attached"
	AttachmentStatusDeleted  AttachmentStatus = "deleted"
)

type Attachment struct {
	ID         string
	MessageID  *string
	UploaderID string
	ChannelID  string
	FileName   string
	MimeType   string
	SizeBytes  int64
	StorageKey string
	Status     AttachmentStatus
	UploadedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}
