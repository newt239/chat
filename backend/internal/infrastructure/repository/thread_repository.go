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
	"github.com/newt239/chat/ent/threadmetadata"
	"github.com/newt239/chat/ent/threadreadstate"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/userthreadfollow"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/ent/workspacemember"
	"github.com/newt239/chat/internal/domain/entity"
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

func (r *threadRepository) FindMetadataByMessageID(ctx context.Context, messageID string) (*entity.ThreadMetadata, error) {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	tm, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithLastReplyUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.ThreadMetadataToEntity(tm), nil
}

func (r *threadRepository) Upsert(ctx context.Context, metadata *entity.ThreadMetadata) error {
	mid, err := utils.ParseUUID(metadata.MessageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Try to find existing
	existing, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		Only(ctx)

	participantIDs := make([]uuid.UUID, len(metadata.ParticipantUserIDs))
	for i, id := range metadata.ParticipantUserIDs {
		participantIDs[i] = utils.ParseUUIDOrNil(id)
	}

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new
			builder := client.ThreadMetadata.Create().
				SetMessageID(mid).
				SetReplyCount(metadata.ReplyCount).
				SetParticipantUserIds(participantIDs)

			if metadata.LastReplyAt != nil {
				builder = builder.SetLastReplyAt(*metadata.LastReplyAt)
			}

			if metadata.LastReplyUserID != nil {
				lruid, err := utils.ParseUUID(*metadata.LastReplyUserID, "last reply user ID")
				if err != nil {
					return err
				}
				builder = builder.SetLastReplyUserID(lruid)
			}

			_, err := builder.Save(ctx)
			if err != nil {
				return err
			}

			// Load edges
			tm, err := client.ThreadMetadata.Query().
				Where(threadmetadata.HasMessageWith(message.ID(mid))).
				WithMessage(func(q *ent.MessageQuery) {
					q.WithChannel(func(q2 *ent.ChannelQuery) {
						q2.WithWorkspace().WithCreatedBy()
					}).WithUser()
				}).
				WithLastReplyUser().
				Only(ctx)
			if err != nil {
				return err
			}

			*metadata = *utils.ThreadMetadataToEntity(tm)
			return nil
		}
		return err
	}

	// Update existing
	builder := client.ThreadMetadata.UpdateOne(existing).
		SetReplyCount(metadata.ReplyCount).
		SetParticipantUserIds(participantIDs)

	if metadata.LastReplyAt != nil {
		builder = builder.SetLastReplyAt(*metadata.LastReplyAt)
	} else {
		builder = builder.ClearLastReplyAt()
	}

	if metadata.LastReplyUserID != nil {
		lruid, err := utils.ParseUUID(*metadata.LastReplyUserID, "last reply user ID")
		if err != nil {
			return err
		}
		builder = builder.SetLastReplyUserID(lruid)
	} else {
		builder = builder.ClearLastReplyUser()
	}

	_, err = builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	tm, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithLastReplyUser().
		Only(ctx)
	if err != nil {
		return err
	}

	*metadata = *utils.ThreadMetadataToEntity(tm)
	return nil
}

func (r *threadRepository) FindMetadataByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.ThreadMetadata, error) {
	if len(messageIDs) == 0 {
		return make(map[string]*entity.ThreadMetadata), nil
	}

	// Parse all message IDs
	parsedIDs := make([]uuid.UUID, 0, len(messageIDs))
	for _, id := range messageIDs {
		parsedID, err := utils.ParseUUID(id, "message ID")
		if err != nil {
			return nil, err
		}
		parsedIDs = append(parsedIDs, parsedID)
	}

	client := transaction.ResolveClient(ctx, r.client)
	metadataList, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.IDIn(parsedIDs...))).
		WithMessage(func(q *ent.MessageQuery) {
			q.WithChannel(func(q2 *ent.ChannelQuery) {
				q2.WithWorkspace().WithCreatedBy()
			}).WithUser()
		}).
		WithLastReplyUser().
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*entity.ThreadMetadata)
	for _, tm := range metadataList {
		messageID := tm.Edges.Message.ID.String()
		result[messageID] = utils.ThreadMetadataToEntity(tm)
	}

	return result, nil
}

func (r *threadRepository) CreateOrUpdateMetadata(ctx context.Context, metadata *entity.ThreadMetadata) error {
	return r.Upsert(ctx, metadata)
}

func (r *threadRepository) IncrementReplyCount(ctx context.Context, messageID string, replyUserID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Try to find existing
	existing, err := client.ThreadMetadata.Query().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new with count 1
			_, err := client.ThreadMetadata.Create().
				SetMessageID(mid).
				SetReplyCount(1).
				SetParticipantUserIds([]uuid.UUID{}).
				Save(ctx)
			return err
		}
		return err
	}

	// Update existing
	return client.ThreadMetadata.UpdateOne(existing).
		SetReplyCount(existing.ReplyCount + 1).
		Exec(ctx)
}

func (r *threadRepository) DeleteMetadata(ctx context.Context, messageID string) error {
	mid, err := utils.ParseUUID(messageID, "message ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.ThreadMetadata.Delete().
		Where(threadmetadata.HasMessageWith(message.ID(mid))).
		Exec(ctx)
	return err
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
		WithUser().
		WithThreadMetadata(func(q *ent.ThreadMetadataQuery) {
			q.WithLastReplyUser()
		})

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
	for i, t := range threads {
		threadIDs[i] = t.ID
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
		if thread.Edges.ThreadMetadata != nil && len(thread.Edges.ThreadMetadata) > 0 {
			metadata := thread.Edges.ThreadMetadata[0]
			replyCount = metadata.ReplyCount
			if !metadata.LastReplyAt.IsZero() {
				lastActivityAt = metadata.LastReplyAt
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
