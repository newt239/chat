package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/channelreadstate"
	"github.com/newt239/chat/ent/message"
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

	// Get the last read time for this user in this channel
	readState, err := client.ChannelReadState.Query().
		Where(
			channelreadstate.HasChannelWith(channel.ID(cid)),
			channelreadstate.HasUserWith(user.ID(uid)),
		).
		Only(ctx)

	var lastReadAt time.Time
	if err != nil {
		if ent.IsNotFound(err) {
			// User has never read this channel, count all messages
			lastReadAt = time.Time{}
		} else {
			return 0, err
		}
	} else {
		lastReadAt = readState.LastReadAt
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
