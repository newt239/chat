package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelreadstate"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type readStateRepository struct {
	client *ent.Client
}

func NewReadStateRepository(client *ent.Client) domainrepository.ReadStateRepository {
	return &readStateRepository{client: client}
}

// 未読数はチャンネルごとに既読位置が異なるため、チャンネルごとの範囲検索を 1 本の SQL にまとめる
// $1: チャンネル ID の配列, $2: ユーザー ID
const unreadCountSQL = `
	SELECT c.id, (
		SELECT COUNT(*) FROM message m
		WHERE m.channel_id = c.id AND m.deleted_at IS NULL
			AND m.created_at > COALESCE(rs.last_read_at, '-infinity')
	)
	FROM unnest($1::uuid[]) AS c(id)
	LEFT JOIN channel_read_state rs ON rs.channel_id = c.id AND rs.user_id = $2`

// 本人へのメンション（グループ経由を含む）を含む未読メッセージをチャンネルごとに数える
const unreadMentionCountSQL = `
	WITH mentioned AS (
		SELECT DISTINCT message_id AS id FROM message_user_mention WHERE user_id = $2
	)
	SELECT m.channel_id, COUNT(*)
	FROM mentioned JOIN message m ON m.id = mentioned.id
	LEFT JOIN channel_read_state rs ON rs.channel_id = m.channel_id AND rs.user_id = $2
	WHERE m.channel_id = ANY($1::uuid[]) AND m.deleted_at IS NULL
		AND m.created_at > COALESCE(rs.last_read_at, '-infinity')
	GROUP BY m.channel_id`

func (r *readStateRepository) Upsert(ctx context.Context, readState *entity.ChannelReadState) error {
	cid, err := utils.ParseUUID(readState.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(readState.UserID, "user ID")
	if err != nil {
		return err
	}

	return transaction.ResolveClient(ctx, r.client).ChannelReadState.Create().
		SetChannelID(cid).
		SetUserID(uid).
		SetLastReadAt(readState.LastReadAt).
		OnConflictColumns(channelreadstate.ChannelColumn, channelreadstate.UserColumn).
		UpdateLastReadAt().
		Exec(ctx)
}

// 既読位置は進めるだけにする。$1: チャンネル ID の配列, $2: ユーザー ID, $3: 既読にした日時
const advanceReadStateSQL = `
	INSERT INTO channel_read_state (id, channel_id, user_id, last_read_at)
	SELECT gen_random_uuid(), c.id, $2, $3 FROM unnest($1::uuid[]) AS c(id)
	ON CONFLICT (channel_id, user_id) DO UPDATE
	SET last_read_at = GREATEST(channel_read_state.last_read_at, EXCLUDED.last_read_at)`

func (r *readStateRepository) AdvanceBatch(ctx context.Context, channelIDs []string, userID string, lastReadAt time.Time) error {
	if len(channelIDs) == 0 {
		return nil
	}
	if _, err := utils.ParseUUIDs(channelIDs, "channel ID"); err != nil {
		return err
	}
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).ExecContext(ctx, advanceReadStateSQL, pq.Array(channelIDs), uid, lastReadAt)
	return err
}

func (r *readStateRepository) GetUnreadCount(ctx context.Context, channelID, userID string) (int, error) {
	counts, err := r.GetUnreadCountBatch(ctx, []string{channelID}, userID)
	if err != nil {
		return 0, err
	}
	return counts[channelID], nil
}

func (r *readStateRepository) GetUnreadCountBatch(ctx context.Context, channelIDs []string, userID string) (map[string]int, error) {
	return r.countByChannel(ctx, unreadCountSQL, channelIDs, userID)
}

func (r *readStateRepository) GetUnreadMentionCountBatch(ctx context.Context, channelIDs []string, userID string) (map[string]int, error) {
	return r.countByChannel(ctx, unreadMentionCountSQL, channelIDs, userID)
}

// countByChannel は (チャンネル ID, 件数) を返す SQL を実行し、件数が 0 より大きいものだけを返します
func (r *readStateRepository) countByChannel(ctx context.Context, query string, channelIDs []string, userID string) (map[string]int, error) {
	result := make(map[string]int, len(channelIDs))
	if len(channelIDs) == 0 {
		return result, nil
	}
	if _, err := utils.ParseUUIDs(channelIDs, "channel ID"); err != nil {
		return nil, err
	}
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, query, pq.Array(channelIDs), uid)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid uuid.UUID
		var count int
		if err := rows.Scan(&cid, &count); err != nil {
			return nil, err
		}
		if count > 0 {
			result[cid.String()] = count
		}
	}
	return result, rows.Err()
}
