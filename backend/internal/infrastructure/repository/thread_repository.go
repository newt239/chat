package repository

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/ent/threadreadstate"
	"github.com/newt239/chat/ent/userthreadfollow"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type threadRepository struct {
	client *ent.Client
}

func NewThreadRepository(client *ent.Client) domainrepository.ThreadRepository {
	return &threadRepository{client: client}
}

// 親ごとの返信数（削除済みを含む）と最新の返信を 1 本の SQL で求める ($1: 親メッセージ ID の配列)
const threadReplySummarySQL = `
	SELECT DISTINCT ON (parent_id) parent_id, COUNT(*) OVER (PARTITION BY parent_id), created_at, user_id
	FROM message WHERE parent_id = ANY($1::uuid[])
	ORDER BY parent_id, created_at DESC`

// CalculateMetadataByMessageIDs は複数のメッセージIDのスレッドメタデータを一括計算します
func (r *threadRepository) CalculateMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*domainrepository.ThreadMetadata, error) {
	result := make(map[string]*domainrepository.ThreadMetadata, len(messageIDs))
	if len(messageIDs) == 0 {
		return result, nil
	}
	parsedIDs, err := parseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]*domainrepository.ThreadMetadata, len(messageIDs))
	for i, id := range messageIDs {
		result[id] = &domainrepository.ThreadMetadata{MessageID: id}
		byID[parsedIDs[i]] = result[id]
	}

	client := transaction.ResolveClient(ctx, r.client)
	rows, err := client.QueryContext(ctx, threadReplySummarySQL, pq.Array(messageIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var parentID, userID uuid.UUID
		var count int
		var lastReplyAt time.Time
		if err := rows.Scan(&parentID, &count, &lastReplyAt, &userID); err != nil {
			return nil, err
		}
		lastReplyUserID := userID.String()
		metadata := byID[parentID]
		metadata.ReplyCount = count
		metadata.LastReplyAt = &lastReplyAt
		metadata.LastReplyUserID = &lastReplyUserID
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

const latestReplyCount = 2

func (r *threadRepository) FindParticipatingThreads(ctx context.Context, input domainrepository.FindParticipatingThreadsInput) (*domainrepository.FindParticipatingThreadsOutput, error) {
	userID, err := parseUUID(input.UserID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	empty := &domainrepository.FindParticipatingThreadsOutput{Items: []domainrepository.ParticipatingThread{}}

	query := client.Message.Query().
		Where(
			message.Not(message.HasParent()),
			message.HasChannelWith(viewableChannel(input.WorkspaceID, userID)),
			message.Or(
				message.HasUserThreadFollowsWith(userthreadfollow.UserID(userID)),
				message.HasRepliesWith(message.UserID(userID)),
				message.HasRepliesWith(message.HasUserMentionsWith(messageusermention.UserID(userID))),
			),
		)

	if cursor := input.Cursor; cursor != nil {
		cursorThreadID, err := parseUUID(cursor.ThreadID, "cursor thread ID")
		if err != nil {
			return nil, err
		}
		query = query.Where(predicate.Message(func(s *sql.Selector) {
			s.Where(sql.P(func(b *sql.Builder) {
				b.WriteString("(" + lastActivityExpr(s) + ", " + s.C(message.FieldID) + ") < (").
					Arg(cursor.LastActivityAt).Comma().Arg(cursorThreadID).WriteString(")")
			}))
		}))
	}

	threads, err := query.
		Order(func(s *sql.Selector) { s.OrderExpr(sql.Expr(lastActivityExpr(s) + " DESC")) }, ent.Desc(message.FieldID)).
		Limit(input.Limit + 1).
		All(ctx)
	if err != nil {
		return nil, err
	}

	hasMore := len(threads) > input.Limit
	if hasMore {
		threads = threads[:input.Limit]
	}
	if len(threads) == 0 {
		return empty, nil
	}

	threadIDs := make([]uuid.UUID, len(threads))
	for i, t := range threads {
		threadIDs[i] = t.ID
	}
	summaries, err := r.summarizeThreads(ctx, threadIDs, userID)
	if err != nil {
		return nil, err
	}

	items := make([]domainrepository.ParticipatingThread, 0, len(threads))
	for _, thread := range threads {
		summary := summaries[thread.ID]
		lastActivityAt := thread.CreatedAt
		if summary.lastReplyAt != nil {
			lastActivityAt = *summary.lastReplyAt
		}
		items = append(items, domainrepository.ParticipatingThread{
			ThreadID:       thread.ID.String(),
			ChannelID:      thread.ChannelID.String(),
			FirstMessage:   messageToEntity(thread),
			LatestReplies:  convertAll(summary.latestReplies, messageToEntity),
			ReplyCount:     summary.replyCount,
			LastActivityAt: lastActivityAt,
			UnreadCount:    summary.unreadCount,
		})
	}

	var nextCursor *domainrepository.ThreadCursor
	if hasMore {
		lastItem := items[len(items)-1]
		nextCursor = &domainrepository.ThreadCursor{LastActivityAt: lastItem.LastActivityAt, ThreadID: lastItem.ThreadID}
	}

	return &domainrepository.FindParticipatingThreadsOutput{Items: items, NextCursor: nextCursor}, nil
}

type threadSummary struct {
	replyCount    int
	unreadCount   int
	lastReplyAt   *time.Time
	latestReplies []*ent.Message
}

// 削除されていない返信の件数・自分以外の未読の件数・最新の返信日時をスレッドごとに集計する ($1: スレッド ID の配列, $2: 閲覧者)
const threadSummarySQL = `
	SELECT r.parent_id, COUNT(*),
		COUNT(*) FILTER (WHERE r.user_id <> $2 AND (s.last_read_at IS NULL OR r.created_at > s.last_read_at)),
		MAX(r.created_at)
	FROM message AS r
	LEFT JOIN thread_read_state AS s ON s.thread_id = r.parent_id AND s.user_id = $2
	WHERE r.parent_id = ANY($1::uuid[]) AND r.deleted_at IS NULL
	GROUP BY r.parent_id`

// スレッドごとに削除されていない最新の返信を $2 件まで選ぶ
const latestRepliesSQL = `
	SELECT id FROM (
		SELECT id, ROW_NUMBER() OVER (PARTITION BY parent_id ORDER BY created_at DESC, id DESC) AS rank
		FROM message WHERE parent_id = ANY($1::uuid[]) AND deleted_at IS NULL
	) AS ranked WHERE rank <= $2`

// summarizeThreads は返信をすべて読み込まずに、一覧に出す件数と最新の返信を SQL で求めます
func (r *threadRepository) summarizeThreads(ctx context.Context, threadIDs []uuid.UUID, userID uuid.UUID) (map[uuid.UUID]*threadSummary, error) {
	client := transaction.ResolveClient(ctx, r.client)
	summaries := make(map[uuid.UUID]*threadSummary, len(threadIDs))
	for _, id := range threadIDs {
		summaries[id] = &threadSummary{}
	}

	rows, err := client.QueryContext(ctx, threadSummarySQL, pq.Array(threadIDs), userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var threadID uuid.UUID
		var lastReplyAt time.Time
		summary := &threadSummary{lastReplyAt: &lastReplyAt}
		if err := rows.Scan(&threadID, &summary.replyCount, &summary.unreadCount, &lastReplyAt); err != nil {
			return nil, err
		}
		summaries[threadID] = summary
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	replyIDs, err := queryUUIDs(ctx, r.client, latestRepliesSQL, pq.Array(threadIDs), latestReplyCount)
	if err != nil {
		return nil, err
	}
	replies, err := client.Message.Query().
		Where(message.IDIn(replyIDs...)).
		Order(ent.Asc(message.FieldCreatedAt), ent.Asc(message.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, reply := range replies {
		summary := summaries[*reply.ParentID]
		summary.latestReplies = append(summary.latestReplies, reply)
	}
	return summaries, nil
}

// lastActivityExpr はスレッドの最終アクティビティ（削除されていない最新の返信、なければ親の投稿日時）を表す SQL 式です
func lastActivityExpr(s *sql.Selector) string {
	return fmt.Sprintf(
		"COALESCE((SELECT MAX(r.%[1]s) FROM %[2]s AS r WHERE r.%[3]s = %[4]s AND r.%[5]s IS NULL), %[6]s)",
		message.FieldCreatedAt, message.Table, message.ParentColumn, s.C(message.FieldID), message.FieldDeletedAt, s.C(message.FieldCreatedAt),
	)
}

func (r *threadRepository) UpsertReadState(ctx context.Context, userID, threadID string, lastReadAt time.Time) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	tid, err := parseUUID(threadID, "thread ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ThreadReadState.Create().
		SetUserID(uid).
		SetThreadID(tid).
		SetLastReadAt(lastReadAt).
		OnConflictColumns(threadreadstate.FieldUserID, threadreadstate.FieldThreadID).
		UpdateLastReadAt().
		Exec(ctx)
}

// FollowThread は既にフォローしていても成功します
func (r *threadRepository) FollowThread(ctx context.Context, userID, threadID string) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	tid, err := parseUUID(threadID, "thread ID")
	if err != nil {
		return err
	}
	err = transaction.ResolveClient(ctx, r.client).UserThreadFollow.Create().
		SetUserID(uid).
		SetThreadID(tid).
		OnConflictColumns(userthreadfollow.FieldUserID, userthreadfollow.FieldThreadID).
		DoNothing().
		Exec(ctx)
	return ignoreConflict(err)
}

func (r *threadRepository) UnfollowThread(ctx context.Context, userID, threadID string) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	tid, err := parseUUID(threadID, "thread ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).UserThreadFollow.Delete().
		Where(userthreadfollow.UserID(uid), userthreadfollow.ThreadID(tid)).
		Exec(ctx)
	return err
}

func (r *threadRepository) FindFollowedThreadIDs(ctx context.Context, userID string, threadIDs []string) (map[string]bool, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	tids, err := parseUUIDs(threadIDs, "thread ID")
	if err != nil {
		return nil, err
	}
	follows, err := transaction.ResolveClient(ctx, r.client).UserThreadFollow.Query().
		Where(userthreadfollow.UserID(uid), userthreadfollow.ThreadIDIn(tids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return idSet(follows, func(f *ent.UserThreadFollow) uuid.UUID { return f.ThreadID }), nil
}

func (r *threadRepository) FindFollowerIDs(ctx context.Context, threadID string) ([]string, error) {
	tid, err := parseUUID(threadID, "thread ID")
	if err != nil {
		return nil, err
	}
	ids, err := transaction.ResolveClient(ctx, r.client).UserThreadFollow.Query().
		Where(userthreadfollow.ThreadID(tid)).
		QueryUser().
		IDs(ctx)
	if err != nil {
		return nil, err
	}
	return uuidStrings(ids), nil
}
