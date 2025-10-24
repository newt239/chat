package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewWorkspaceRepository instead.
func NewWorkspaceRepository(db *gorm.DB) domainrepository.WorkspaceRepository {
	return persistence.NewWorkspaceRepository(db)
}
