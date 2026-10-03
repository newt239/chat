package service

import (
	"context"
	"time"
)

// StorageService は添付ファイルなどを置くストレージの署名付き URL を発行します
type StorageService interface {
	// GenerateUploadURL は expires が 0 なら StorageConfig の既定の期限で発行します
	GenerateUploadURL(ctx context.Context, storageKey, mimeType string, sizeBytes int64, expires time.Duration) (string, error)
	// GenerateDownloadURL は expires が 0 なら StorageConfig の既定の期限で発行します
	GenerateDownloadURL(ctx context.Context, storageKey string, expires time.Duration) (string, error)
	DeleteObject(ctx context.Context, storageKey string) error
}

// StorageConfig はアップロードの上限と署名付き URL の既定の期限です
type StorageConfig interface {
	GetMaxFileSize() int64
	GetUploadExpires() time.Duration
	GetDownloadExpires() time.Duration
}
