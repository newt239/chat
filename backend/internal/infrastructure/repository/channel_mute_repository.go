package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelmute"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type channelMuteRepository struct {
	client *ent.Client
}

func NewChannelMuteRepository(client *ent.Client) domainrepository.ChannelMuteRepository {
	return &channelMuteRepository{client: client}
}

func (r *channelMuteRepository) SetMuted(ctx context.Context, userID string, channelID string, muted bool) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	if !muted {
		_, err = client.ChannelMute.Delete().
			Where(channelmute.UserID(uid), channelmute.ChannelID(cid)).
			Exec(ctx)
		return err
	}
	// ミュート済みでも結果は同じなので成功とみなす
	err = client.ChannelMute.Create().
		SetUserID(uid).
		SetChannelID(cid).
		OnConflictColumns(channelmute.FieldUserID, channelmute.FieldChannelID).
		DoNothing().
		Exec(ctx)
	return ignoreConflict(err)
}

func (r *channelMuteRepository) FindMutedChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	cids, err := utils.ParseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	mutes, err := transaction.ResolveClient(ctx, r.client).ChannelMute.Query().
		Where(channelmute.UserID(uid), channelmute.ChannelIDIn(cids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(mutes))
	for _, m := range mutes {
		result[m.ChannelID.String()] = true
	}
	return result, nil
}

func (r *channelMuteRepository) FindMutedUserIDs(ctx context.Context, channelID string, userIDs []string) (map[string]bool, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	uids, err := utils.ParseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}
	mutes, err := transaction.ResolveClient(ctx, r.client).ChannelMute.Query().
		Where(channelmute.ChannelID(cid), channelmute.UserIDIn(uids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(mutes))
	for _, m := range mutes {
		result[m.UserID.String()] = true
	}
	return result, nil
}
