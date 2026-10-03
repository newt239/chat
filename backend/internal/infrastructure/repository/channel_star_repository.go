package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelstar"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type channelStarRepository struct {
	client *ent.Client
}

func NewChannelStarRepository(client *ent.Client) domainrepository.ChannelStarRepository {
	return &channelStarRepository{client: client}
}

func (r *channelStarRepository) SetStarred(ctx context.Context, userID string, channelID string, starred bool) error {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	cid, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	if !starred {
		_, err = client.ChannelStar.Delete().Where(channelstar.UserID(uid), channelstar.ChannelID(cid)).Exec(ctx)
		return err
	}
	return ignoreConflict(client.ChannelStar.Create().
		SetUserID(uid).
		SetChannelID(cid).
		OnConflictColumns(channelstar.FieldUserID, channelstar.FieldChannelID).
		DoNothing().
		Exec(ctx))
}

func (r *channelStarRepository) FindStarredChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	cids, err := parseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	stars, err := transaction.ResolveClient(ctx, r.client).ChannelStar.Query().
		Where(channelstar.UserID(uid), channelstar.ChannelIDIn(cids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return idSet(stars, func(s *ent.ChannelStar) uuid.UUID { return s.ChannelID }), nil
}
