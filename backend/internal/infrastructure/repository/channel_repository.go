package repository

import (
	"context"
	"strings"

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
		SetIsPrivate(ch.IsPrivate).
		SetChannelType(string(ch.Type))

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

func (r *channelRepository) SearchAccessibleChannels(ctx context.Context, workspaceID, userID string, query string, limit int, offset int) ([]*entity.Channel, int, error) {
	wID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, 0, err
	}

	uID, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, 0, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	trimmedQuery := strings.TrimSpace(query)

	channelQuery := client.Channel.Query().
		Where(
			channel.HasWorkspaceWith(workspace.ID(wID)),
			channel.HasMembersWith(channelmember.HasUserWith(user.ID(uID))),
		)

	if trimmedQuery != "" {
		channelQuery = channelQuery.Where(
			channel.Or(
				channel.NameContainsFold(trimmedQuery),
				channel.DescriptionContainsFold(trimmedQuery),
			),
		)
	}

	total, err := channelQuery.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if offset > 0 {
		channelQuery = channelQuery.Offset(offset)
	}

	if limit > 0 {
		channelQuery = channelQuery.Limit(limit)
	}

	channels, err := channelQuery.
		WithWorkspace(func(q *ent.WorkspaceQuery) {
			q.WithCreatedBy()
		}).
		WithCreatedBy().
		Order(ent.Asc(channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*entity.Channel, 0, len(channels))
	for _, c := range channels {
		result = append(result, utils.ChannelToEntity(c))
	}

	return result, total, nil
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

func (r *channelRepository) FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error) {
	wID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	uid1, err := utils.ParseUUID(userID1, "user ID 1")
	if err != nil {
		return nil, err
	}

	uid2, err := utils.ParseUUID(userID2, "user ID 2")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)

	existingChannels, err := client.Channel.Query().
		Where(
			channel.HasWorkspaceWith(workspace.ID(wID)),
			channel.ChannelTypeEQ("dm"),
			channel.HasMembersWith(channelmember.HasUserWith(user.ID(uid1))),
			channel.HasMembersWith(channelmember.HasUserWith(user.ID(uid2))),
		).
		WithWorkspace().
		WithCreatedBy().
		WithMembers(func(q *ent.ChannelMemberQuery) {
			q.WithUser()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range existingChannels {
		if len(ch.Edges.Members) == 2 {
			return utils.ChannelToEntity(ch), nil
		}
	}

	dmChannel := &entity.Channel{
		WorkspaceID: workspaceID,
		Name:        "dm_" + uid1.String() + "_" + uid2.String(),
		IsPrivate:   true,
		Type:        entity.ChannelTypeDM,
		CreatedBy:   userID1,
	}

	if err := r.Create(ctx, dmChannel); err != nil {
		return nil, err
	}

	return dmChannel, nil
}

func (r *channelRepository) FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string, name string) (*entity.Channel, error) {
	wID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	cID, err := utils.ParseUUID(creatorID, "creator ID")
	if err != nil {
		return nil, err
	}

	if len(memberIDs) > 9 {
		return nil, entity.ErrGroupDMMaxMembers
	}

	client := transaction.ResolveClient(ctx, r.client)

	existingChannels, err := client.Channel.Query().
		Where(
			channel.HasWorkspaceWith(workspace.ID(wID)),
			channel.ChannelTypeEQ("group_dm"),
			channel.HasCreatedByWith(user.ID(cID)),
		).
		WithWorkspace().
		WithCreatedBy().
		WithMembers(func(q *ent.ChannelMemberQuery) {
			q.WithUser()
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range existingChannels {
		if len(ch.Edges.Members) == len(memberIDs) {
			memberMap := make(map[string]bool)
			for _, m := range ch.Edges.Members {
				if m.Edges.User != nil {
					memberMap[m.Edges.User.ID.String()] = true
				}
			}

			allMatch := true
			for _, mid := range memberIDs {
				if !memberMap[mid] {
					allMatch = false
					break
				}
			}

			if allMatch {
				return utils.ChannelToEntity(ch), nil
			}
		}
	}

	channelName := name
	if channelName == "" {
		channelName = "group_dm_" + cID.String()
	}

	groupDMChannel := &entity.Channel{
		WorkspaceID: workspaceID,
		Name:        channelName,
		IsPrivate:   true,
		Type:        entity.ChannelTypeGroupDM,
		CreatedBy:   creatorID,
	}

	if err := r.Create(ctx, groupDMChannel); err != nil {
		return nil, err
	}

	return groupDMChannel, nil
}

func (r *channelRepository) FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
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
			channel.ChannelTypeIn("dm", "group_dm"),
			channel.HasMembersWith(channelmember.HasUserWith(user.ID(uID))),
		).
		WithWorkspace(func(q *ent.WorkspaceQuery) {
			q.WithCreatedBy()
		}).
		WithCreatedBy().
		WithMembers(func(q *ent.ChannelMemberQuery) {
			q.WithUser()
		}).
		Order(ent.Desc(channel.FieldUpdatedAt)).
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
