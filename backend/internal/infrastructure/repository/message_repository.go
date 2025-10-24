package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewMessageRepository instead.
func NewMessageRepository(db *gorm.DB) domainrepository.MessageRepository {
	return persistence.NewMessageRepository(db)
}
