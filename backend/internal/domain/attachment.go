package domain

import "time"

type Attachment struct {
	ID         string
	MessageID  string
	FileName   string
	MimeType   string
	SizeBytes  int64
	StorageKey string
	CreatedAt  time.Time
}

type AttachmentRepository interface {
	FindByID(id string) (*Attachment, error)
	FindByMessageID(messageID string) ([]*Attachment, error)
	Create(attachment *Attachment) error
	Delete(id string) error
}
