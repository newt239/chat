package service

import (
	"context"
	"time"
)

const (
	MaxUploadSize      = 1 << 30
	UploadURLExpires   = 15 * time.Minute
	DownloadURLExpires = 5 * time.Minute
)

// StorageService は添付ファイルなどを置くストレージの署名付き URL を発行します
type StorageService interface {
	GenerateUploadURL(ctx context.Context, storageKey, mimeType string, expires time.Duration) (string, error)
	GenerateDownloadURL(ctx context.Context, storageKey string, expires time.Duration) (string, error)
	DeleteObject(ctx context.Context, storageKey string) error
}
