package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewSessionRepository instead.
func NewSessionRepository(db *gorm.DB) domainrepository.SessionRepository {
	return persistence.NewSessionRepository(db)
}
