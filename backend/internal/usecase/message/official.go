package message

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// ensureNotOfficial は公式アプリの投稿を編集・削除させません
func ensureNotOfficial(ctx context.Context, userRepo domainrepository.UserRepository, message *entity.Message) error {
	author, err := userRepo.FindByID(ctx, message.UserID)
	if err != nil {
		return fmt.Errorf("failed to load author: %w", err)
	}
	if author != nil && author.IsOfficial {
		return ErrOfficialMessage
	}
	return nil
}
