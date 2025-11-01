package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/channelreadstate"
	"github.com/newt239/chat/ent/message"
	"github.com/newt239/chat/ent/messagegroupmention"
	"github.com/newt239/chat/ent/messageusermention"
	"github.com/newt239/chat/ent/usergroup"
	"github.com/newt239/chat/ent/usergroupmember"
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

func (r *readStateRepository) Upsert(ctx context.Context, readState *entity.ChannelReadState) error {
	cid, err := utils.ParseUUID(readState.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(readState.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Try to find existing
	existing, err := client.ChannelReadState.Query().
		Where(
			channelreadstate.HasChannelWith(channel.ID(cid)),
			channelreadstate.HasUserWith(user.ID(uid)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Create new
			_, err := client.ChannelReadState.Create().
				SetChannelID(cid).
				SetUserID(uid).
				SetLastReadAt(readState.LastReadAt).
				Save(ctx)
			if err != nil {
				return err
			}

			// Load edges
			crs, err := client.ChannelReadState.Query().
				Where(
					channelreadstate.HasChannelWith(channel.ID(cid)),
					channelreadstate.HasUserWith(user.ID(uid)),
				).
				WithChannel(func(q *ent.ChannelQuery) {
					q.WithWorkspace().WithCreatedBy()
				}).
				WithUser().
				Only(ctx)
			if err != nil {
				return err
			}

			*readState = *utils.ChannelReadStateToEntity(crs)
			return nil
		}
		return err
	}

	// Update existing
	_, err = client.ChannelReadState.UpdateOne(existing).
		SetLastReadAt(readState.LastReadAt).
		Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	crs, err := client.ChannelReadState.Query().
		Where(
			channelreadstate.HasChannelWith(channel.ID(cid)),
			channelreadstate.HasUserWith(user.ID(uid)),
		).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		Only(ctx)
	if err != nil {
		return err
	}

	*readState = *utils.ChannelReadStateToEntity(crs)
	return nil
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
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.ChannelReadStateToEntity(crs), nil
}

func (r *readStateRepository) UpdateLastReadAt(ctx context.Context, channelID, userID string, lastReadAt time.Time) error {
	readState := &entity.ChannelReadState{
		ChannelID:  channelID,
		UserID:     userID,
		LastReadAt: lastReadAt,
	}
	return r.Upsert(ctx, readState)
}

// getLastReadAt は指定されたチャネルとユーザーの既読時刻を取得します
func (r *readStateRepository) getLastReadAt(ctx context.Context, client *ent.Client, cid, uid uuid.UUID) (time.Time, error) {
	readState, err := client.ChannelReadState.Query().
		Where(
			channelreadstate.HasChannelWith(channel.ID(cid)),
			channelreadstate.HasUserWith(user.ID(uid)),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}

	return readState.LastReadAt, nil
}

func (r *readStateRepository) GetUnreadCount(ctx context.Context, channelID, userID string) (int, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return 0, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return 0, err
	}

	client := transaction.ResolveClient(ctx, r.client)

	lastReadAt, err := r.getLastReadAt(ctx, client, cid, uid)
	if err != nil {
		return 0, err
	}

	// Count messages created after the last read time
	count, err := client.Message.Query().
		Where(
			message.HasChannelWith(channel.ID(cid)),
			message.CreatedAtGT(lastReadAt),
			message.DeletedAtIsNil(),
		).
		Count(ctx)

	return count, err
}

func (r *readStateRepository) GetUnreadChannels(ctx context.Context, userID string) (map[string]int, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Get all channels the user is a member of
	channels, err := client.Channel.Query().
		Where(channel.HasMembersWith(channelmember.HasUserWith(user.ID(uid)))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, ch := range channels {
		count, err := r.GetUnreadCount(ctx, ch.ID.String(), userID)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			result[ch.ID.String()] = count
		}
	}

	return result, nil
}

func (r *readStateRepository) GetUnreadMentionCount(ctx context.Context, channelID, userID string) (int, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return 0, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return 0, err
	}

	client := transaction.ResolveClient(ctx, r.client)

	lastReadAt, err := r.getLastReadAt(ctx, client, cid, uid)
	if err != nil {
		return 0, err
	}

	// Get user's group IDs
	userGroups, err := client.UserGroup.Query().
		Where(usergroup.HasMembersWith(usergroupmember.HasUserWith(user.ID(uid)))).
		All(ctx)
	if err != nil {
		return 0, err
	}

	groupIDs := make([]uuid.UUID, len(userGroups))
	for i, g := range userGroups {
		groupIDs[i] = g.ID
	}

	// Count unique messages with mentions after lastReadAt
	// 1. User mentions: messages where the user is mentioned
	userMentionMessages, err := client.MessageUserMention.Query().
		Where(
			messageusermention.HasUserWith(user.ID(uid)),
			messageusermention.HasMessageWith(
				message.HasChannelWith(channel.ID(cid)),
				message.CreatedAtGT(lastReadAt),
				message.DeletedAtIsNil(),
			),
		).
		QueryMessage().
		IDs(ctx)
	if err != nil {
		return 0, err
	}

	// 2. Group mentions: messages where user's groups are mentioned
	var groupMentionMessages []uuid.UUID
	if len(groupIDs) > 0 {
		groupMentionMessages, err = client.MessageGroupMention.Query().
			Where(
				messagegroupmention.HasGroupWith(usergroup.IDIn(groupIDs...)),
				messagegroupmention.HasMessageWith(
					message.HasChannelWith(channel.ID(cid)),
					message.CreatedAtGT(lastReadAt),
					message.DeletedAtIsNil(),
				),
			).
			QueryMessage().
			IDs(ctx)
		if err != nil {
			return 0, err
		}
	}

	// Combine and deduplicate message IDs
	messageIDSet := make(map[uuid.UUID]bool)
	for _, id := range userMentionMessages {
		messageIDSet[id] = true
	}
	for _, id := range groupMentionMessages {
		messageIDSet[id] = true
	}

	return len(messageIDSet), nil
}

func (r *readStateRepository) GetUnreadMentionCountBatch(ctx context.Context, channelIDs []string, userID string) (map[string]int, error) {
	if len(channelIDs) == 0 {
		return make(map[string]int), nil
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	cids := make([]uuid.UUID, 0, len(channelIDs))
	for _, cid := range channelIDs {
		parsed, err := utils.ParseUUID(cid, "channel ID")
		if err != nil {
			return nil, err
		}
		cids = append(cids, parsed)
	}

	client := transaction.ResolveClient(ctx, r.client)

	// Get all read states for these channels
	readStates, err := client.ChannelReadState.Query().
		Where(
			channelreadstate.HasChannelWith(channel.IDIn(cids...)),
			channelreadstate.HasUserWith(user.ID(uid)),
		).
		WithChannel().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Map channelID -> lastReadAt
	lastReadAtMap := make(map[uuid.UUID]time.Time)
	for _, rs := range readStates {
		if rs.Edges.Channel != nil {
			lastReadAtMap[rs.Edges.Channel.ID] = rs.LastReadAt
		}
	}

	// Get user's group IDs
	userGroups, err := client.UserGroup.Query().
		Where(usergroup.HasMembersWith(usergroupmember.HasUserWith(user.ID(uid)))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]uuid.UUID, len(userGroups))
	for i, g := range userGroups {
		groupIDs[i] = g.ID
	}

	result := make(map[string]int)

	for _, cid := range cids {
		lastReadAt := lastReadAtMap[cid] // Zero value if not found

		// User mentions
		userMentionMessages, err := client.MessageUserMention.Query().
			Where(
				messageusermention.HasUserWith(user.ID(uid)),
				messageusermention.HasMessageWith(
					message.HasChannelWith(channel.ID(cid)),
					message.CreatedAtGT(lastReadAt),
					message.DeletedAtIsNil(),
				),
			).
			QueryMessage().
			IDs(ctx)
		if err != nil {
			return nil, err
		}

		// Group mentions
		var groupMentionMessages []uuid.UUID
		if len(groupIDs) > 0 {
			groupMentionMessages, err = client.MessageGroupMention.Query().
				Where(
					messagegroupmention.HasGroupWith(usergroup.IDIn(groupIDs...)),
					messagegroupmention.HasMessageWith(
						message.HasChannelWith(channel.ID(cid)),
						message.CreatedAtGT(lastReadAt),
						message.DeletedAtIsNil(),
					),
				).
				QueryMessage().
				IDs(ctx)
			if err != nil {
				return nil, err
			}
		}

		// Deduplicate
		messageIDSet := make(map[uuid.UUID]bool)
		for _, id := range userMentionMessages {
			messageIDSet[id] = true
		}
		for _, id := range groupMentionMessages {
			messageIDSet[id] = true
		}

		count := len(messageIDSet)
		if count > 0 {
			result[cid.String()] = count
		}
	}

	return result, nil
}
