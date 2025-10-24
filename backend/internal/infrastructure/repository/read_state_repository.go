package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewReadStateRepository instead.
func NewReadStateRepository(db *gorm.DB) domainrepository.ReadStateRepository {
	return persistence.NewReadStateRepository(db)
}
