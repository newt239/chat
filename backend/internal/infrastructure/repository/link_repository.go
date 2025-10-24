package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewMessageLinkRepository instead.
func NewMessageLinkRepository(db *gorm.DB) domainrepository.MessageLinkRepository {
	return persistence.NewMessageLinkRepository(db)
}
