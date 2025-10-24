package entity

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
