package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelstar"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type channelStarRepository struct {
	client *ent.Client
}

func NewChannelStarRepository(client *ent.Client) domainrepository.ChannelStarRepository {
	return &channelStarRepository{client: client}
}

func (r *channelStarRepository) SetStarred(ctx context.Context, userID string, channelID string, starred bool) error {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	if !starred {
		_, err = client.ChannelStar.Delete().
			Where(channelstar.UserID(uid), channelstar.ChannelID(cid)).
			Exec(ctx)
		return err
	}
	// 付与済みでも結果は同じなので成功とみなす
	err = client.ChannelStar.Create().
		SetUserID(uid).
		SetChannelID(cid).
		OnConflictColumns(channelstar.FieldUserID, channelstar.FieldChannelID).
		DoNothing().
		Exec(ctx)
	return ignoreConflict(err)
}

func (r *channelStarRepository) FindStarredChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error) {
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
	starredIDs, err := client.ChannelStar.Query().
		Where(channelstar.UserID(uid), channelstar.ChannelIDIn(cids...)).
		QueryChannel().
		IDs(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(starredIDs))
	for _, id := range starredIDs {
		result[id.String()] = true
	}
	return result, nil
}
