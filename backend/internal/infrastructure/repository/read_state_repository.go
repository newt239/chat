package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelreadstate"
	"github.com/newt239/chat/ent/user"
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
		SELECT COUNT(*) FROM messages m
		WHERE m.message_channel = c.id AND m.deleted_at IS NULL
			AND m.created_at > COALESCE(rs.last_read_at, '-infinity')
	)
	FROM unnest($1::uuid[]) AS c(id)
	LEFT JOIN channel_read_states rs ON rs.channel_read_state_channel = c.id AND rs.channel_read_state_user = $2`

// 本人か所属グループへのメンションを含む未読メッセージをチャンネルごとに数える
const unreadMentionCountSQL = `
	WITH mentioned AS (
		SELECT message_user_mention_message AS id FROM message_user_mentions WHERE message_user_mention_user = $2
		UNION
		SELECT gm.message_group_mention_message FROM message_group_mentions gm
		JOIN user_group_members ugm ON ugm.user_group_member_group = gm.message_group_mention_group
		WHERE ugm.user_group_member_user = $2
	)
	SELECT m.message_channel, COUNT(*)
	FROM mentioned JOIN messages m ON m.id = mentioned.id
	LEFT JOIN channel_read_states rs ON rs.channel_read_state_channel = m.message_channel AND rs.channel_read_state_user = $2
	WHERE m.message_channel = ANY($1::uuid[]) AND m.deleted_at IS NULL
		AND m.created_at > COALESCE(rs.last_read_at, '-infinity')
	GROUP BY m.message_channel`

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

func (r *readStateRepository) FindByChannelAndUser(ctx context.Context, channelID, userID string) (*entity.ChannelReadState, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	crs, err := client.ChannelReadState.Query().
		Where(
			channelreadstate.HasChannelWith(channel.ID(cid)),
			channelreadstate.HasUserWith(user.ID(uid)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &entity.ChannelReadState{ChannelID: channelID, UserID: userID, LastReadAt: crs.LastReadAt}, nil
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
	if _, err := parseUUIDs(channelIDs, "channel ID"); err != nil {
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
