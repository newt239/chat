package repository

import (
	"context"

	"github.com/google/uuid"

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
	err = client.ChannelMute.Create().
		SetUserID(uid).
		SetChannelID(cid).
		Exec(ctx)
	// ミュート済みの場合は一意制約違反になるが、結果は同じなので成功とみなす
	if ent.IsConstraintError(err) {
		return nil
	}
	return err
}

func (r *channelMuteRepository) FindMutedChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	cids := make([]uuid.UUID, 0, len(channelIDs))
	for _, id := range channelIDs {
		cid, err := utils.ParseUUID(id, "channel ID")
		if err != nil {
			return nil, err
		}
		cids = append(cids, cid)
	}

	client := transaction.ResolveClient(ctx, r.client)
	mutedIDs, err := client.ChannelMute.Query().
		Where(channelmute.UserID(uid), channelmute.ChannelIDIn(cids...)).
		QueryChannel().
		IDs(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(mutedIDs))
	for _, id := range mutedIDs {
		result[id.String()] = true
	}
	return result, nil
}
