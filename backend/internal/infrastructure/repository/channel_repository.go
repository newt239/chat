package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewChannelRepository instead.
func NewChannelRepository(db *gorm.DB) domainrepository.ChannelRepository {
	return persistence.NewChannelRepository(db)
}
