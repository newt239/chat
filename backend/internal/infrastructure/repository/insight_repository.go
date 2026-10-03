package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

// 集計は日付の区切りやタイムゾーンを扱うため生 SQL で書く
type insightRepository struct {
	client *ent.Client
}

func NewInsightRepository(client *ent.Client) domainrepository.InsightRepository {
	return &insightRepository{client: client}
}

// $1: ワークスペース ID, $2: 期間の開始, $3: 期間の終了
const workspaceMessagesSQL = `
	SELECT m.user_id AS user_id, m.created_at
	FROM message m JOIN channel c ON c.id = m.channel_id
	WHERE c.workspace_id = $1 AND m.deleted_at IS NULL AND m.created_at >= $2 AND m.created_at < $3`

// 投稿とリアクションを「活動」とみなす
const workspaceActivitiesSQL = `
	SELECT user_id, created_at, 1 AS is_message FROM (` + workspaceMessagesSQL + `) msg
	UNION ALL
	SELECT r.user_id, r.created_at, 0
	FROM message_reaction r
	JOIN message m ON m.id = r.message_id
	JOIN channel c ON c.id = m.channel_id
	WHERE c.workspace_id = $1 AND r.created_at >= $2 AND r.created_at < $3`

const workspaceAttachmentsSQL = `
	FROM attachment a JOIN channel c ON c.id = a.channel_id
	WHERE c.workspace_id = $1 AND a.status = 'attached'`

func (r *insightRepository) query(ctx context.Context, query string, args []any, scan func(*sql.Rows) error) error {
	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *insightRepository) queryInt(ctx context.Context, query string, args ...any) (int64, error) {
	var n int64
	err := r.query(ctx, query, args, func(rows *sql.Rows) error { return rows.Scan(&n) })
	return n, err
}

func (r *insightRepository) CountMessages(ctx context.Context, workspaceID string, from, to time.Time) (int, error) {
	n, err := r.queryInt(ctx, `SELECT COUNT(*) FROM (`+workspaceMessagesSQL+`) t`, workspaceID, from, to)
	return int(n), err
}

func (r *insightRepository) CountActiveMembers(ctx context.Context, workspaceID string, from, to time.Time) (int, error) {
	n, err := r.queryInt(ctx, `SELECT COUNT(DISTINCT user_id) FROM (`+workspaceActivitiesSQL+`) t`, workspaceID, from, to)
	return int(n), err
}

func (r *insightRepository) CountMembersJoinedBefore(ctx context.Context, workspaceID string, before time.Time) (int, error) {
	n, err := r.queryInt(ctx, `
		SELECT COUNT(*) FROM workspace_member
		WHERE workspace_id = $1 AND joined_at < $2 AND suspended_at IS NULL`, workspaceID, before)
	return int(n), err
}

func (r *insightRepository) StorageBytesBefore(ctx context.Context, workspaceID string, before time.Time) (int64, error) {
	return r.queryInt(ctx, `SELECT COALESCE(SUM(a.size_bytes), 0) `+workspaceAttachmentsSQL+` AND a.created_at < $2`, workspaceID, before)
}

func (r *insightRepository) StorageByMimeType(ctx context.Context, workspaceID string) ([]entity.MimeTypeUsage, error) {
	var result []entity.MimeTypeUsage
	err := r.query(ctx, `SELECT a.mime_type, SUM(a.size_bytes), COUNT(*) `+workspaceAttachmentsSQL+` GROUP BY a.mime_type`,
		[]any{workspaceID},
		func(rows *sql.Rows) error {
			var u entity.MimeTypeUsage
			if err := rows.Scan(&u.MimeType, &u.Bytes, &u.FileCount); err != nil {
				return err
			}
			result = append(result, u)
			return nil
		})
	return result, err
}

func (r *insightRepository) DailyActivity(ctx context.Context, workspaceID string, from, to time.Time, loc *time.Location) ([]entity.DailyActivity, error) {
	return r.daily(ctx, `
		SELECT to_char(created_at AT TIME ZONE $4, 'YYYY-MM-DD') AS d, SUM(is_message), COUNT(DISTINCT user_id)
		FROM (`+workspaceActivitiesSQL+`) t GROUP BY d`,
		workspaceID, from, to, loc.String())
}

func (r *insightRepository) DailyMessageCountsByUser(ctx context.Context, workspaceID, userID string, from, to time.Time, loc *time.Location) ([]entity.DailyActivity, error) {
	return r.daily(ctx, `
		SELECT to_char(created_at AT TIME ZONE $4, 'YYYY-MM-DD') AS d, COUNT(*), 1
		FROM (`+workspaceMessagesSQL+`) t WHERE user_id = $5 GROUP BY d`,
		workspaceID, from, to, loc.String(), userID)
}

func (r *insightRepository) daily(ctx context.Context, query string, args ...any) ([]entity.DailyActivity, error) {
	var result []entity.DailyActivity
	err := r.query(ctx, query, args, func(rows *sql.Rows) error {
		var d entity.DailyActivity
		if err := rows.Scan(&d.Date, &d.MessageCount, &d.ActiveMemberCount); err != nil {
			return err
		}
		result = append(result, d)
		return nil
	})
	return result, err
}

func (r *insightRepository) ChannelMessageCounts(ctx context.Context, workspaceID string, viewerID *string, from, recentFrom, to time.Time) ([]entity.ChannelMessageCount, error) {
	var viewer any
	if viewerID != nil {
		viewer = *viewerID
	}
	var result []entity.ChannelMessageCount
	err := r.query(ctx, `
		SELECT c.id, c.name, c.channel_type <> 'public',
			COUNT(m.id),
			COUNT(m.id) FILTER (WHERE m.created_at >= $3)
		FROM channel c
		LEFT JOIN message m ON m.channel_id = c.id AND m.deleted_at IS NULL AND m.created_at >= $2 AND m.created_at < $4
		WHERE c.workspace_id = $1
			AND COALESCE(c.channel_type, 'public') IN ('public', 'private')
			AND ($5::uuid IS NULL OR c.channel_type = 'public' OR EXISTS (
				SELECT 1 FROM channel_member cm WHERE cm.channel_id = c.id AND cm.user_id = $5::uuid
			))
		GROUP BY c.id, c.name, c.channel_type
		ORDER BY COUNT(m.id) DESC, c.name`,
		[]any{workspaceID, from, recentFrom, to, viewer},
		func(rows *sql.Rows) error {
			var c entity.ChannelMessageCount
			if err := rows.Scan(&c.ChannelID, &c.Name, &c.IsPrivate, &c.MessageCount, &c.RecentMessageCount); err != nil {
				return err
			}
			result = append(result, c)
			return nil
		})
	return result, err
}

func (r *insightRepository) Heatmap(ctx context.Context, workspaceID string, from, to time.Time, loc *time.Location) ([]entity.HeatmapCell, error) {
	var result []entity.HeatmapCell
	err := r.query(ctx, `
		SELECT EXTRACT(ISODOW FROM created_at AT TIME ZONE $4)::int AS dow, EXTRACT(HOUR FROM created_at AT TIME ZONE $4)::int AS hour, COUNT(*)
		FROM (`+workspaceMessagesSQL+`) t GROUP BY dow, hour`,
		[]any{workspaceID, from, to, loc.String()},
		func(rows *sql.Rows) error {
			var c entity.HeatmapCell
			if err := rows.Scan(&c.Weekday, &c.Hour, &c.MessageCount); err != nil {
				return err
			}
			result = append(result, c)
			return nil
		})
	return result, err
}

func (r *insightRepository) MemberActivities(ctx context.Context, workspaceID string, from, to time.Time) ([]entity.MemberActivity, error) {
	var result []entity.MemberActivity
	err := r.query(ctx, `
		SELECT wm.user_id,
			(SELECT COUNT(*) FROM (`+workspaceMessagesSQL+`) t WHERE t.user_id = wm.user_id),
			(SELECT COALESCE(SUM(a.size_bytes), 0) `+workspaceAttachmentsSQL+` AND a.uploader_id = wm.user_id),
			(SELECT MAX(m.created_at) FROM message m JOIN channel c ON c.id = m.channel_id
				WHERE c.workspace_id = $1 AND m.deleted_at IS NULL AND m.user_id = wm.user_id)
		FROM workspace_member wm WHERE wm.workspace_id = $1`,
		[]any{workspaceID, from, to},
		func(rows *sql.Rows) error {
			var a entity.MemberActivity
			var last sql.NullTime
			if err := rows.Scan(&a.UserID, &a.MessageCount, &a.StorageBytes, &last); err != nil {
				return err
			}
			if last.Valid {
				a.LastMessageAt = &last.Time
			}
			result = append(result, a)
			return nil
		})
	return result, err
}
