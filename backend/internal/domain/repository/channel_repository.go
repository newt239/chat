package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type ChannelRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Channel, error)
	FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error)
	FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	// FindBrowsableChannels は参加の有無を問わず閲覧できるアーカイブされていないチャンネル（公開と参加中の非公開）を返します
	FindBrowsableChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	// FindLastMessageAtBatch はスレッドの返信を除いた最後のメッセージの投稿日時を返します。メッセージのないチャンネルは含みません
	FindLastMessageAtBatch(ctx context.Context, channelIDs []string) (map[string]time.Time, error)
	CountMembersBatch(ctx context.Context, channelIDs []string) (map[string]int, error)
	SearchAccessibleChannels(ctx context.Context, workspaceID string, userID string, query string, limit int, offset int) ([]*entity.Channel, int, error)
	Create(ctx context.Context, channel *entity.Channel) error
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id string) error
	FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error)
	FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string, name string) (*entity.Channel, error)
	FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	FindByNames(ctx context.Context, workspaceID string, names []string) ([]*entity.Channel, error)
	// FindDescendants はパスの前方一致で子孫チャンネルを返します
	FindDescendants(ctx context.Context, ch *entity.Channel) ([]*entity.Channel, error)
}
