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
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/userthreadfollow"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/ent/workspacemember"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type threadRepository struct {
	client *ent.Client
}

func NewThreadRepository(client *ent.Client) domainrepository.ThreadRepository {
	return &threadRepository{client: client}
}

// CalculateMetadataByMessageID は指定されたメッセージIDのスレッドメタデータを計算します
func (r *threadRepository) CalculateMetadataByMessageID(ctx context.Context, messageID string) (*domainrepository.ThreadMetadata, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}
	exists, err := transaction.ResolveClient(ctx, r.client).Message.Query().Where(message.ID(mid)).Exist(ctx)
	if err != nil || !exists {
		return nil, err
	}
	metadata, err := r.CalculateMetadataByMessageIDs(ctx, []string{messageID})
	if err != nil {
		return nil, err
	}
	return metadata[messageID], nil
}

// 親ごとの返信数（削除済みを含む）と最新の返信を 1 本の SQL で求める
// $1: 親メッセージ ID の配列
const threadReplySummarySQL = `
	SELECT DISTINCT ON (message_parent) message_parent, COUNT(*) OVER (PARTITION BY message_parent), created_at, message_user
	FROM messages WHERE message_parent = ANY($1::uuid[])
	ORDER BY message_parent, created_at DESC`

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
		result[id] = &domainrepository.ThreadMetadata{MessageID: id, ParticipantUserIDs: []string{}}
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

	follows, err := client.UserThreadFollow.Query().
		Where(userthreadfollow.HasThreadWith(message.IDIn(parsedIDs...))).
		WithUser(func(q *ent.UserQuery) { q.Select(user.FieldID) }).
		WithThread(func(q *ent.MessageQuery) { q.Select(message.FieldID) }).
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, follow := range follows {
		metadata := byID[follow.Edges.Thread.ID]
		metadata.ParticipantUserIDs = append(metadata.ParticipantUserIDs, follow.Edges.User.ID.String())
	}

	return result, nil
}

const latestReplyCount = 2

func (r *threadRepository) FindParticipatingThreads(ctx context.Context, input domainrepository.FindParticipatingThreadsInput) (*domainrepository.FindParticipatingThreadsOutput, error) {
	userID, err := utils.ParseUUID(input.UserID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	empty := &domainrepository.FindParticipatingThreadsOutput{Items: []domainrepository.ParticipatingThread{}}

	isMember, err := client.WorkspaceMember.Query().
		Where(
			workspacemember.HasUserWith(user.ID(userID)),
			workspacemember.HasWorkspaceWith(workspace.ID(input.WorkspaceID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return empty, nil
	}

	query := client.Message.Query().
		Where(
			message.Not(message.HasParent()),
			message.HasChannelWith(viewableChannel(input.WorkspaceID, userID)),
			message.Or(
				message.HasUserThreadFollowsWith(userthreadfollow.HasUserWith(user.ID(userID))),
				message.HasRepliesWith(message.HasUserWith(user.ID(userID))),
				message.HasRepliesWith(message.HasUserMentionsWith(messageusermention.HasUserWith(user.ID(userID)))),
			),
		)

	if input.CursorLastActivityAt != nil && input.CursorThreadID != nil {
		cursorThreadID, err := utils.ParseUUID(*input.CursorThreadID, "cursor thread ID")
		if err != nil {
			return nil, err
		}
		query = query.Where(predicate.Message(func(s *sql.Selector) {
			s.Where(sql.P(func(b *sql.Builder) {
				b.WriteString("(" + lastActivityExpr(s) + ", " + s.C(message.FieldID) + ") < (").
					Arg(*input.CursorLastActivityAt).Comma().Arg(cursorThreadID).WriteString(")")
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

	replies, err := client.Message.Query().
		Where(message.ParentIDIn(threadIDs...), message.DeletedAtIsNil()).
		Order(ent.Asc(message.FieldCreatedAt), ent.Asc(message.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	repliesByThread := make(map[uuid.UUID][]*ent.Message)
	for _, reply := range replies {
		repliesByThread[*reply.ParentID] = append(repliesByThread[*reply.ParentID], reply)
	}

	readStates, err := client.ThreadReadState.Query().
		Where(
			threadreadstate.HasUserWith(user.ID(userID)),
			threadreadstate.HasThreadWith(message.IDIn(threadIDs...)),
		).
		WithThread(func(q *ent.MessageQuery) { q.Select(message.FieldID) }).
		All(ctx)
	if err != nil {
		return nil, err
	}
	lastReadAt := make(map[uuid.UUID]time.Time)
	for _, rs := range readStates {
		lastReadAt[rs.Edges.Thread.ID] = rs.LastReadAt
	}

	items := make([]domainrepository.ParticipatingThread, 0, len(threads))
	for _, thread := range threads {
		threadReplies := repliesByThread[thread.ID]
		lastActivityAt := thread.CreatedAt
		if len(threadReplies) > 0 {
			lastActivityAt = threadReplies[len(threadReplies)-1].CreatedAt
		}

		readAt, hasRead := lastReadAt[thread.ID]
		unreadCount := 0
		for _, reply := range threadReplies {
			if reply.UserID != userID && (!hasRead || reply.CreatedAt.After(readAt)) {
				unreadCount++
			}
		}

		channelID := thread.ChannelID.String()

		items = append(items, domainrepository.ParticipatingThread{
			ThreadID:       thread.ID.String(),
			ChannelID:      &channelID,
			FirstMessage:   utils.MessageToEntity(thread),
			LatestReplies:  toMessageEntities(threadReplies[max(len(threadReplies)-latestReplyCount, 0):]),
			ReplyCount:     len(threadReplies),
			LastActivityAt: lastActivityAt,
			UnreadCount:    unreadCount,
		})
	}

	var nextCursor *domainrepository.ThreadCursor
	if hasMore {
		lastItem := items[len(items)-1]
		nextCursor = &domainrepository.ThreadCursor{LastActivityAt: lastItem.LastActivityAt, ThreadID: lastItem.ThreadID}
	}

	return &domainrepository.FindParticipatingThreadsOutput{Items: items, NextCursor: nextCursor}, nil
}

// lastActivityExpr はスレッドの最終アクティビティ（削除されていない最新の返信、なければ親の投稿日時）を表す SQL 式です
func lastActivityExpr(s *sql.Selector) string {
	return fmt.Sprintf(
		"COALESCE((SELECT MAX(r.%[1]s) FROM %[2]s AS r WHERE r.%[3]s = %[4]s AND r.%[5]s IS NULL), %[6]s)",
		message.FieldCreatedAt, message.Table, message.ParentColumn, s.C(message.FieldID), message.FieldDeletedAt, s.C(message.FieldCreatedAt),
	)
}

func (r *threadRepository) UpsertReadState(ctx context.Context, userID, threadID string, lastReadAt time.Time) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	tid, err := utils.ParseUUID(threadID, "thread ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// 既存のreadstateを探す
	existing, err := client.ThreadReadState.Query().
		Where(
			threadreadstate.HasUserWith(user.ID(uid)),
			threadreadstate.HasThreadWith(message.ID(tid)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// 新規作成
			_, err := client.ThreadReadState.Create().
				SetUserID(uid).
				SetThreadID(tid).
				SetLastReadAt(lastReadAt).
				Save(ctx)
			return err
		}
		return err
	}

	// 更新
	return client.ThreadReadState.UpdateOne(existing).
		SetLastReadAt(lastReadAt).
		Exec(ctx)
}

func (r *threadRepository) GetReadState(ctx context.Context, userID, threadID string) (*time.Time, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	tid, err := utils.ParseUUID(threadID, "thread ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	readState, err := client.ThreadReadState.Query().
		Where(
			threadreadstate.HasUserWith(user.ID(uid)),
			threadreadstate.HasThreadWith(message.ID(tid)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &readState.LastReadAt, nil
}

func (r *threadRepository) FollowThread(ctx context.Context, userID, threadID string) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	tid, err := utils.ParseUUID(threadID, "thread ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// 既に存在するかチェック
	exists, err := client.UserThreadFollow.Query().
		Where(
			userthreadfollow.HasUserWith(user.ID(uid)),
			userthreadfollow.HasThreadWith(message.ID(tid)),
		).
		Exist(ctx)

	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = client.UserThreadFollow.Create().
		SetUserID(uid).
		SetThreadID(tid).
		Save(ctx)

	return err
}

func (r *threadRepository) UnfollowThread(ctx context.Context, userID, threadID string) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	tid, err := utils.ParseUUID(threadID, "thread ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	_, err = client.UserThreadFollow.Delete().
		Where(
			userthreadfollow.HasUserWith(user.ID(uid)),
			userthreadfollow.HasThreadWith(message.ID(tid)),
		).
		Exec(ctx)

	return err
}

func (r *threadRepository) IsFollowing(ctx context.Context, userID, threadID string) (bool, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return false, err
	}
	tid, err := utils.ParseUUID(threadID, "thread ID")
	if err != nil {
		return false, err
	}

	client := transaction.ResolveClient(ctx, r.client)

	return client.UserThreadFollow.Query().
		Where(
			userthreadfollow.HasUserWith(user.ID(uid)),
			userthreadfollow.HasThreadWith(message.ID(tid)),
		).
		Exist(ctx)
}
