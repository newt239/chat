package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewAttachmentRepository instead.
func NewAttachmentRepository(db *gorm.DB) domainrepository.AttachmentRepository {
	return persistence.NewAttachmentRepository(db)
}
