package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewUserGroupRepository instead.
func NewUserGroupRepository(db *gorm.DB) domainrepository.UserGroupRepository {
	return persistence.NewUserGroupRepository(db)
}
