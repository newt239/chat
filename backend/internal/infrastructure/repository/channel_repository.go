package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/ent/workspace"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type channelRepository struct {
	client *ent.Client
}

func NewChannelRepository(client *ent.Client) domainrepository.ChannelRepository {
	return &channelRepository{client: client}
}

func (r *channelRepository) FindByID(ctx context.Context, id string) (*entity.Channel, error) {
	channelID, err := utils.ParseUUID(id, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	c, err := client.Channel.Query().
		Where(channel.ID(channelID)).
		WithWorkspace().
		WithCreatedBy().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.ChannelToEntity(c), nil
}

func (r *channelRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.Channel, error) {
	wid, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(channel.HasWorkspaceWith(workspace.ID(wid))).
		WithWorkspace().
		WithCreatedBy().
		Order(ent.Asc(channel.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Channel, 0, len(channels))
	for _, c := range channels {
		result = append(result, utils.ChannelToEntity(c))
	}

	return result, nil
}

func (r *channelRepository) FindByWorkspaceIDAndUserID(ctx context.Context, workspaceID, userID string, includePrivate bool) ([]*entity.Channel, error) {
	wid, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	query := client.Channel.Query().
		Where(channel.HasWorkspaceWith(workspace.ID(wid)))

	if includePrivate {
		query = query.Where(
			channel.Or(
				channel.IsPrivate(false),
				channel.HasMembersWith(channelmember.HasUserWith(user.ID(uid))),
			),
		)
	} else {
		query = query.Where(channel.IsPrivate(false))
	}

	channels, err := query.
		WithWorkspace().
		WithCreatedBy().
		Order(ent.Asc(channel.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Channel, 0, len(channels))
	for _, c := range channels {
		result = append(result, utils.ChannelToEntity(c))
	}

	return result, nil
}

func (r *channelRepository) Create(ctx context.Context, ch *entity.Channel) error {
	wid, err := utils.ParseUUID(ch.WorkspaceID, "workspace ID")
	if err != nil {
		return err
	}

	createdBy, err := utils.ParseUUID(ch.CreatedBy, "created_by user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Channel.Create().
		SetWorkspaceID(wid).
		SetCreatedByID(createdBy).
		SetName(ch.Name).
		SetIsPrivate(ch.IsPrivate)

	if ch.ID != "" {
		channelID, err := utils.ParseUUID(ch.ID, "channel ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(channelID)
	}

	if ch.Description != nil {
		builder = builder.SetDescription(*ch.Description)
	}

	c, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	c, err = client.Channel.Query().
		Where(channel.ID(c.ID)).
		WithWorkspace().
		WithCreatedBy().
		Only(ctx)
	if err != nil {
		return err
	}

	*ch = *utils.ChannelToEntity(c)
	return nil
}

func (r *channelRepository) Update(ctx context.Context, ch *entity.Channel) error {
	channelID, err := utils.ParseUUID(ch.ID, "channel ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Channel.UpdateOneID(channelID).
		SetName(ch.Name)

	if ch.Description != nil {
		builder = builder.SetDescription(*ch.Description)
	} else {
		builder = builder.ClearDescription()
	}

	c, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	c, err = client.Channel.Query().
		Where(channel.ID(c.ID)).
		WithWorkspace().
		WithCreatedBy().
		Only(ctx)
	if err != nil {
		return err
	}

	ch.UpdatedAt = c.UpdatedAt
	return nil
}

func (r *channelRepository) Delete(ctx context.Context, id string) error {
	channelID, err := utils.ParseUUID(id, "channel ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.Channel.DeleteOneID(channelID).Exec(ctx)
}

func (r *channelRepository) FindAccessibleChannels(ctx context.Context, workspaceID, userID string) ([]*entity.Channel, error) {
	wID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	uID, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(
			channel.HasWorkspaceWith(workspace.ID(wID)),
			channel.HasMembersWith(channelmember.HasUserWith(user.ID(uID))),
		).
		WithWorkspace(func(q *ent.WorkspaceQuery) {
			q.WithCreatedBy()
		}).
		WithCreatedBy().
		Order(ent.Asc(channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.Channel, 0, len(channels))
	for _, c := range channels {
		result = append(result, utils.ChannelToEntity(c))
	}

	return result, nil
}
