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
	// List は新しい順に limit 件返します。cursor は前のページの nextCursor で、それより前にピン留めしたものを返します
	List(ctx context.Context, channelID string, limit int, cursor *string) (pins []*entity.MessagePin, nextCursor *string, err error)
	// FindByMessageIDs は Message を設定しません
	FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.MessagePin, error)
}
