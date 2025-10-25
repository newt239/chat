package service

// StorageService defines the interface for storage operations
type StorageService interface {
	GenerateUploadURL(storageKey, mimeType string, sizeBytes int64, expiresIn interface{}) (string, error)
	GenerateDownloadURL(storageKey string, expiresIn interface{}) (string, error)
}

// StorageConfig defines the configuration for storage operations
type StorageConfig interface {
	GetMaxFileSize() int64
	GetUploadExpires() interface{}
	GetDownloadExpires() interface{}
}
