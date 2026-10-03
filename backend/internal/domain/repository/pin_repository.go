package repository

import (
	"context"

	"github.com/newt239/chat/internal/domain/entity"
)

type PinRepository interface {
	// Create は同じメッセージが既にピン留めされていれば ErrPinExists を返します
	Create(ctx context.Context, pin *entity.MessagePin) error
	// Delete はピンがなくても成功します
	Delete(ctx context.Context, channelID, messageID string) error
	// List はピン留めした新しい順に limit 件返します
	List(ctx context.Context, channelID string, limit int) ([]*entity.MessagePin, error)
	// FindByMessageIDs は Message を設定しません
	FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.MessagePin, error)
}
