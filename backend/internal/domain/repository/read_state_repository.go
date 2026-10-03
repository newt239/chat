package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ReadStateRepository interface {
	Upsert(ctx context.Context, readState *entity.ChannelReadState) error
	// AdvanceBatch は既読位置を lastReadAt まで進めます。既にそれより後まで読んでいるチャンネルは戻さない
	AdvanceBatch(ctx context.Context, channelIDs []string, userID string, lastReadAt time.Time) error
	GetUnreadCountBatch(ctx context.Context, channelIDs []string, userID string) (map[string]int, error)
	GetUnreadMentionCountBatch(ctx context.Context, channelIDs []string, userID string) (map[string]int, error)
}
