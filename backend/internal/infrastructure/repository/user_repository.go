package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewUserRepository instead.
func NewUserRepository(db *gorm.DB) domainrepository.UserRepository {
	return persistence.NewUserRepository(db)
}
