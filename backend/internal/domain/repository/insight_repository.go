package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// InsightRepository はワークスペースの指標を集計します。期間はすべて [from, to) です
type InsightRepository interface {
	CountMessages(ctx context.Context, workspaceID string, from, to time.Time) (int, error)
	// CountActiveMembers はメッセージの投稿か既読の更新をしたメンバーの人数を返します
	CountActiveMembers(ctx context.Context, workspaceID string, from, to time.Time) (int, error)
	CountMembersJoinedBefore(ctx context.Context, workspaceID string, before time.Time) (int, error)
	StorageBytesBefore(ctx context.Context, workspaceID string, before time.Time) (int64, error)
	StorageByMimeType(ctx context.Context, workspaceID string) ([]entity.MimeTypeUsage, error)
	DailyActivity(ctx context.Context, workspaceID string, from, to time.Time, loc *time.Location) ([]entity.DailyActivity, error)
	DailyMessageCountsByUser(ctx context.Context, workspaceID, userID string, from, to time.Time, loc *time.Location) ([]entity.DailyActivity, error)
	// ChannelMessageCounts は DM を除くチャンネルを返します。viewerID を渡すとその人が見られない非公開チャンネルを除きます
	ChannelMessageCounts(ctx context.Context, workspaceID string, viewerID *string, from, recentFrom, to time.Time) ([]entity.ChannelMessageCount, error)
	Heatmap(ctx context.Context, workspaceID string, from, to time.Time, loc *time.Location) ([]entity.HeatmapCell, error)
	MemberActivities(ctx context.Context, workspaceID string, from, to time.Time) ([]entity.MemberActivity, error)
}
