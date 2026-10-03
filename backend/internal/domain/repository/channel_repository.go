package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type BrowsableChannelMembership int

const (
	BrowsableChannelMembershipAll BrowsableChannelMembership = iota
	BrowsableChannelMembershipJoined
	BrowsableChannelMembershipNotJoined
)

type BrowsableChannelSort int

const (
	BrowsableChannelSortName BrowsableChannelSort = iota
	BrowsableChannelSortMemberCount
)

type BrowsableChannelFilter struct {
	// 名前と説明の部分一致。空ならすべて
	Query      string
	Membership BrowsableChannelMembership
	Sort       BrowsableChannelSort
	Limit      int
	Offset     int
}

type ChannelRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Channel, error)
	FindAccessibleChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	// FindBrowsableChannels は参加の有無を問わず閲覧できるチャンネル（公開と参加中の非公開）を返します
	FindBrowsableChannels(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	// SearchBrowsableChannels は FindBrowsableChannels と同じ範囲を条件で絞り込み、1 ページ分と総数を返します
	SearchBrowsableChannels(ctx context.Context, workspaceID string, userID string, filter BrowsableChannelFilter) ([]*entity.Channel, int, error)
	// FindLastMessageAtBatch はスレッドの返信を除いた最後のメッセージの投稿日時を返します。メッセージのないチャンネルは含みません
	FindLastMessageAtBatch(ctx context.Context, channelIDs []string) (map[string]time.Time, error)
	CountMembersBatch(ctx context.Context, channelIDs []string) (map[string]int, error)
	Create(ctx context.Context, channel *entity.Channel) error
	Update(ctx context.Context, channel *entity.Channel) error
	// FindOrCreateDM と FindOrCreateGroupDM は DM を返し、参加者を全員参加させます
	FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error)
	FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string, name string) (*entity.Channel, error)
	FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error)
	FindByNames(ctx context.Context, workspaceID string, names []string) ([]*entity.Channel, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.Channel, error)
	// FindDescendants はパスの前方一致で子孫チャンネルを返します
	FindDescendants(ctx context.Context, ch *entity.Channel) ([]*entity.Channel, error)
}
