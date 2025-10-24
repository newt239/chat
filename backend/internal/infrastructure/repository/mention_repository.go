package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

// Deprecated: use persistence.NewMessageUserMentionRepository instead.
func NewMessageUserMentionRepository(db *gorm.DB) domainrepository.MessageUserMentionRepository {
	return persistence.NewMessageUserMentionRepository(db)
}

// Deprecated: use persistence.NewMessageGroupMentionRepository instead.
func NewMessageGroupMentionRepository(db *gorm.DB) domainrepository.MessageGroupMentionRepository {
	return persistence.NewMessageGroupMentionRepository(db)
}
