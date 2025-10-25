package repository

import (
	"github.com/example/chat/internal/adapter/gateway/persistence"
	domainrepository "github.com/example/chat/internal/domain/repository"
	"gorm.io/gorm"
)

func NewThreadRepository(db *gorm.DB) domainrepository.ThreadRepository {
	return persistence.NewThreadRepository(db)
}
