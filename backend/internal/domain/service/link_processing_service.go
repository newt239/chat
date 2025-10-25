package service

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

// LinkProcessingService defines the interface for link processing operations
type LinkProcessingService interface {
	ProcessLinks(ctx context.Context, body string) ([]*entity.MessageLink, error)
}
