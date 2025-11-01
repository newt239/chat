package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messageusermention"
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

	client := transaction.ResolveClient(ctx, r.client)

	// メッセージの存在確認
	exists, err := client.Message.Query().Where(message.ID(mid)).Exist(ctx)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	// 返信を取得
	replies, err := client.Message.Query().
		Where(message.HasParentWith(message.ID(mid))).
		Order(ent.Desc(message.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// メタデータを計算
	replyCount := len(replies)
	var lastReplyAt *time.Time
	var lastReplyUserID *string

	if replyCount > 0 {
		lastReply := replies[0]
		lastReplyAt = &lastReply.CreatedAt
		userID := lastReply.Edges.User.ID.String()
		lastReplyUserID = &userID
	}

	// 参加者を取得（UserThreadFollowから）
	follows, err := client.UserThreadFollow.Query().
		Where(userthreadfollow.HasThreadWith(message.ID(mid))).
		WithUser().
		All(ctx)
	if err != nil {
		return nil, err
	}

	participantUserIDs := make([]string, 0, len(follows))
	for _, follow := range follows {
		if follow.Edges.User != nil {
			participantUserIDs = append(participantUserIDs, follow.Edges.User.ID.String())
		}
	}

	return &domainrepository.ThreadMetadata{
		MessageID:          messageID,
		ReplyCount:         replyCount,
		LastReplyAt:        lastReplyAt,
		LastReplyUserID:    lastReplyUserID,
		ParticipantUserIDs: participantUserIDs,
	}, nil
}

// CalculateMetadataByMessageIDs は複数のメッセージIDのスレッドメタデータを一括計算します
func (r *threadRepository) CalculateMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*domainrepository.ThreadMetadata, error) {
	if len(messageIDs) == 0 {
		return make(map[string]*domainrepository.ThreadMetadata), nil
	}

	// Parse all message IDs
	parsedIDs := make([]uuid.UUID, 0, len(messageIDs))
	idStrMap := make(map[uuid.UUID]string)
	for _, id := range messageIDs {
		parsedID, err := utils.ParseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		parsedIDs = append(parsedIDs, parsedID)
		idStrMap[parsedID] = id
	}

	client := transaction.ResolveClient(ctx, r.client)

	// 全ての返信を取得
	replies, err := client.Message.Query().
		Where(message.HasParentWith(message.IDIn(parsedIDs...))).
		WithUser().
		Order(ent.Desc(message.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 親メッセージごとにグループ化
	repliesByParent := make(map[uuid.UUID][]*ent.Message)
	for _, reply := range replies {
		parentEdges := reply.QueryParent().IDsX(ctx)
		if len(parentEdges) > 0 {
			parentID := parentEdges[0]
			repliesByParent[parentID] = append(repliesByParent[parentID], reply)
		}
	}

	// 全てのフォローを取得
	follows, err := client.UserThreadFollow.Query().
		Where(userthreadfollow.HasThreadWith(message.IDIn(parsedIDs...))).
		WithUser().
		WithThread().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// スレッドごとにフォロワーをグループ化
	followersByThread := make(map[uuid.UUID][]string)
	for _, follow := range follows {
		if follow.Edges.Thread != nil && follow.Edges.User != nil {
			threadID := follow.Edges.Thread.ID
			userID := follow.Edges.User.ID.String()
			followersByThread[threadID] = append(followersByThread[threadID], userID)
		}
	}

	// 結果を構築
	result := make(map[string]*domainrepository.ThreadMetadata)
	for _, parsedID := range parsedIDs {
		messageID := idStrMap[parsedID]
		threadReplies := repliesByParent[parsedID]
		replyCount := len(threadReplies)

		var lastReplyAt *time.Time
		var lastReplyUserID *string

		if replyCount > 0 {
			lastReply := threadReplies[0]
			lastReplyAt = &lastReply.CreatedAt
			if lastReply.Edges.User != nil {
				userID := lastReply.Edges.User.ID.String()
				lastReplyUserID = &userID
			}
		}

		participantUserIDs := followersByThread[parsedID]
		if participantUserIDs == nil {
			participantUserIDs = []string{}
		}

		result[messageID] = &domainrepository.ThreadMetadata{
			MessageID:          messageID,
			ReplyCount:         replyCount,
			LastReplyAt:        lastReplyAt,
			LastReplyUserID:    lastReplyUserID,
			ParticipantUserIDs: participantUserIDs,
		}
	}

	return result, nil
}

func (r *threadRepository) FindParticipatingThreads(ctx context.Context, input domainrepository.FindParticipatingThreadsInput) (*domainrepository.FindParticipatingThreadsOutput, error) {
	userID, err := utils.ParseUUID(input.UserID, "user ID")
	if err != nil {
		return nil, err
	}
	workspaceID, err := utils.ParseUUID(input.WorkspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// 参加しているワークスペースとチャンネルのIDを取得
	workspaceMember, err := client.WorkspaceMember.Query().
		Where(
			workspacemember.HasUserWith(user.ID(userID)),
			workspacemember.HasWorkspaceWith(workspace.ID(workspaceID)),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &domainrepository.FindParticipatingThreadsOutput{
				Items:      []domainrepository.ParticipatingThread{},
				NextCursor: nil,
			}, nil
		}
		return nil, err
	}
	if workspaceMember == nil {
		return &domainrepository.FindParticipatingThreadsOutput{
			Items:      []domainrepository.ParticipatingThread{},
			NextCursor: nil,
		}, nil
	}

	// ユーザーが参加しているチャンネルIDを取得
	channelMembers, err := client.ChannelMember.Query().
		Where(channelmember.HasUserWith(user.ID(userID))).
		WithChannel().
		All(ctx)
	if err != nil {
		return nil, err
	}

	accessibleChannelIDs := make([]uuid.UUID, 0, len(channelMembers))
	for _, cm := range channelMembers {
		accessibleChannelIDs = append(accessibleChannelIDs, cm.Edges.Channel.ID)
	}

	// スレッド起点メッセージ（parent_id == null）で、参加中のものを検索
	query := client.Message.Query().
		Where(
			message.Not(message.HasParent()),
			message.HasChannelWith(channel.IDIn(accessibleChannelIDs...)),
		)

	// 参加条件でフィルタ
	query = query.Where(
		message.Or(
			// フォロー中
			message.HasUserThreadFollowsWith(userthreadfollow.HasUserWith(user.ID(userID))),
			// 返信した
			message.HasRepliesWith(message.HasUserWith(user.ID(userID))),
			// メンションされた
			message.HasRepliesWith(message.HasUserMentionsWith(messageusermention.HasUserWith(user.ID(userID)))),
		),
	)

	// カーソルベースのページネーション
	if input.CursorLastActivityAt != nil && input.CursorThreadID != nil {
		cursorThreadID, err := utils.ParseUUID(*input.CursorThreadID, "cursor thread ID")
		if err != nil {
			return nil, err
		}
		query = query.Where(
			message.Or(
				message.And(
					message.CreatedAtLT(*input.CursorLastActivityAt),
				),
				message.And(
					message.CreatedAtEQ(*input.CursorLastActivityAt),
					message.IDLT(cursorThreadID),
				),
			),
		)
	}

	// ソート: lastActivityAt DESC, threadId DESC
	query = query.
		Order(ent.Desc(message.FieldCreatedAt), ent.Desc(message.FieldID)).
		Limit(input.Limit + 1). // 次のページがあるか確認するため+1
		WithChannel().
		WithUser()

	threads, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	// 次のページがあるかチェック
	hasMore := len(threads) > input.Limit
	if hasMore {
		threads = threads[:input.Limit]
	}

	// スレッドIDリストを作成
	threadIDs := make([]uuid.UUID, len(threads))
	threadIDStrs := make([]string, len(threads))
	for i, t := range threads {
		threadIDs[i] = t.ID
		threadIDStrs[i] = t.ID.String()
	}

	// 各スレッドのメタデータを計算
	metadataMap, err := r.CalculateMetadataByMessageIDs(ctx, threadIDStrs)
	if err != nil {
		return nil, err
	}

	// 各スレッドの未読数を計算
	readStates, err := client.ThreadReadState.Query().
		Where(
			threadreadstate.HasUserWith(user.ID(userID)),
			threadreadstate.HasThreadWith(message.IDIn(threadIDs...)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	readStateMap := make(map[uuid.UUID]time.Time)
	for _, rs := range readStates {
		threadID, err := rs.QueryThread().OnlyID(ctx)
		if err != nil {
			continue
		}
		readStateMap[threadID] = rs.LastReadAt
	}

	// 結果を構築
	items := make([]domainrepository.ParticipatingThread, 0, len(threads))
	for _, thread := range threads {
		var channelID *string
		if thread.Edges.Channel != nil {
			cid := thread.Edges.Channel.ID.String()
			channelID = &cid
		}

		// スレッドメタデータから情報取得
		replyCount := 0
		lastActivityAt := thread.CreatedAt
		if metadata, ok := metadataMap[thread.ID.String()]; ok {
			replyCount = metadata.ReplyCount
			if metadata.LastReplyAt != nil {
				lastActivityAt = *metadata.LastReplyAt
			}
		}

		// 未読数を計算
		unreadCount := 0
		if lastReadAt, ok := readStateMap[thread.ID]; ok {
			count, err := client.Message.Query().
				Where(
					message.HasParentWith(message.ID(thread.ID)),
					message.CreatedAtGT(lastReadAt),
				).
				Count(ctx)
			if err == nil {
				unreadCount = count
			}
		} else {
			// 既読状態がない場合は全て未読
			count, err := client.Message.Query().
				Where(message.HasParentWith(message.ID(thread.ID))).
				Count(ctx)
			if err == nil {
				unreadCount = count
			}
		}

		firstMessage := utils.MessageToEntity(thread)

		items = append(items, domainrepository.ParticipatingThread{
			ThreadID:       thread.ID.String(),
			ChannelID:      channelID,
			FirstMessage:   firstMessage,
			ReplyCount:     replyCount,
			LastActivityAt: lastActivityAt,
			UnreadCount:    unreadCount,
		})
	}

	var nextCursor *domainrepository.ThreadCursor
	if hasMore && len(items) > 0 {
		lastItem := items[len(items)-1]
		nextCursor = &domainrepository.ThreadCursor{
			LastActivityAt: lastItem.LastActivityAt,
			ThreadID:       lastItem.ThreadID,
		}
	}

	return &domainrepository.FindParticipatingThreadsOutput{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
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
