package repository

import (
	"context"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
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
	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(channel.WorkspaceID(workspaceID)).
		Order(ent.Asc(channel.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return channelsToEntities(channels), nil
}

func (r *channelRepository) FindBrowsableChannels(ctx context.Context, workspaceID, userID string) ([]*entity.Channel, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(
			channel.WorkspaceID(workspaceID),
			channel.ChannelTypeIn(string(entity.ChannelTypePublic), string(entity.ChannelTypePrivate)),
			channel.ArchivedAtIsNil(),
			channel.Or(
				channel.ChannelType(string(entity.ChannelTypePublic)),
				channel.HasMembersWith(channelmember.UserID(uid)),
			),
		).
		Order(ent.Asc(channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return channelsToEntities(channels), nil
}

func (r *channelRepository) SearchBrowsableChannels(ctx context.Context, workspaceID, userID string, filter domainrepository.BrowsableChannelFilter) ([]*entity.Channel, int, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, 0, err
	}

	isMember := channel.HasMembersWith(channelmember.UserID(uid))
	query := transaction.ResolveClient(ctx, r.client).Channel.Query().
		Where(
			channel.WorkspaceID(workspaceID),
			channel.ChannelTypeIn(string(entity.ChannelTypePublic), string(entity.ChannelTypePrivate)),
			channel.ArchivedAtIsNil(),
			channel.Or(channel.ChannelType(string(entity.ChannelTypePublic)), isMember),
		)
	if keyword := strings.TrimSpace(filter.Query); keyword != "" {
		query = query.Where(channel.Or(channel.NameContainsFold(keyword), channel.DescriptionContainsFold(keyword)))
	}
	switch filter.Membership {
	case domainrepository.BrowsableChannelMembershipJoined:
		query = query.Where(isMember)
	case domainrepository.BrowsableChannelMembershipNotJoined:
		query = query.Where(channel.Not(isMember))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	order := []channel.OrderOption{channel.ByName()}
	if filter.Sort == domainrepository.BrowsableChannelSortMemberCount {
		order = append([]channel.OrderOption{channel.ByMembersCount(sql.OrderDesc())}, order...)
	}
	channels, err := query.
		Order(order...).
		Limit(filter.Limit).
		Offset(filter.Offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return channelsToEntities(channels), total, nil
}

// スレッドの返信と削除済みを除いた最後のメッセージの投稿日時
const lastMessageAtSQL = `
	SELECT channel_id, MAX(created_at) FROM message
	WHERE channel_id = ANY($1::uuid[]) AND parent_id IS NULL AND deleted_at IS NULL
	GROUP BY channel_id`

func (r *channelRepository) FindLastMessageAtBatch(ctx context.Context, channelIDs []string) (map[string]time.Time, error) {
	result := make(map[string]time.Time, len(channelIDs))
	if len(channelIDs) == 0 {
		return result, nil
	}
	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, lastMessageAtSQL, pq.Array(channelIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid uuid.UUID
		var at time.Time
		if err := rows.Scan(&cid, &at); err != nil {
			return nil, err
		}
		result[cid.String()] = at
	}
	return result, rows.Err()
}

const memberCountSQL = `
	SELECT channel_id, COUNT(*) FROM channel_member
	WHERE channel_id = ANY($1::uuid[])
	GROUP BY channel_id`

func (r *channelRepository) CountMembersBatch(ctx context.Context, channelIDs []string) (map[string]int, error) {
	result := make(map[string]int, len(channelIDs))
	if len(channelIDs) == 0 {
		return result, nil
	}
	rows, err := transaction.ResolveClient(ctx, r.client).QueryContext(ctx, memberCountSQL, pq.Array(channelIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid uuid.UUID
		var count int
		if err := rows.Scan(&cid, &count); err != nil {
			return nil, err
		}
		result[cid.String()] = count
	}
	return result, rows.Err()
}

func (r *channelRepository) Create(ctx context.Context, ch *entity.Channel) error {
	// workspaceID is slug (string)
	createdBy, err := utils.ParseUUID(ch.CreatedBy, "created_by user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.Channel.Create().
		SetWorkspaceID(ch.WorkspaceID).
		SetCreatedByID(createdBy).
		SetName(ch.Name).
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

	if ch.ParentID != nil {
		parentID, err := utils.ParseUUID(*ch.ParentID, "parent channel ID")
		if err != nil {
			return err
		}
		builder = builder.SetParentID(parentID)
	}

	c, err := builder.Save(ctx)
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

	builder = builder.SetChannelType(string(ch.Type)).SetNillableArchivedAt(ch.ArchivedAt)
	if ch.ArchivedAt == nil {
		builder = builder.ClearArchivedAt()
	}

	c, err := builder.Save(ctx)
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
	uID, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, 0, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	trimmedQuery := strings.TrimSpace(query)

	channelQuery := client.Channel.Query().
		Where(
			channel.WorkspaceID(workspaceID),
			channel.HasMembersWith(channelmember.UserID(uID)),
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
		Order(ent.Asc(channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return channelsToEntities(channels), total, nil
}

func (r *channelRepository) FindAccessibleChannels(ctx context.Context, workspaceID, userID string) ([]*entity.Channel, error) {
	uID, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(
			channel.WorkspaceID(workspaceID),
			channel.HasMembersWith(channelmember.UserID(uID)),
			// DM・グループ DM は ListDirectMessages で返す
			channel.ChannelTypeIn(string(entity.ChannelTypePublic), string(entity.ChannelTypePrivate)),
		).
		Order(ent.Asc(channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return channelsToEntities(channels), nil
}

// FindOrCreateDM は 2 人の DM を返します。なければ作ります。同時に作られても dm_key の一意制約で 1 つにまとまる
func (r *channelRepository) FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error) {
	return r.findOrCreateByDMKey(ctx, &entity.Channel{
		WorkspaceID: workspaceID,
		Name:        "dm_" + userID1 + "_" + userID2,
		Type:        entity.ChannelTypeDM,
		CreatedBy:   userID1,
	}, dmKey("dm:", userID1, userID2))
}

// FindOrCreateGroupDM はメンバーがまったく同じグループ DM を返します。なければ name で作ります
func (r *channelRepository) FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string, name string) (*entity.Channel, error) {
	if len(memberIDs) > entity.MaxGroupDMMembers {
		return nil, entity.ErrGroupDMMaxMembers
	}
	if name == "" {
		name = "group_dm_" + creatorID
	}
	return r.findOrCreateByDMKey(ctx, &entity.Channel{
		WorkspaceID: workspaceID,
		Name:        name,
		Type:        entity.ChannelTypeGroupDM,
		CreatedBy:   creatorID,
	}, dmKey("g:", memberIDs...))
}

// dmKey は参加者の ID を並べ替えてつなげ、参加者が同じ DM を同じキーにします
func dmKey(prefix string, userIDs ...string) string {
	sorted := slices.Clone(userIDs)
	slices.Sort(sorted)
	return prefix + strings.Join(slices.Compact(sorted), ",")
}

func (r *channelRepository) findOrCreateByDMKey(ctx context.Context, ch *entity.Channel, key string) (*entity.Channel, error) {
	createdBy, err := utils.ParseUUID(ch.CreatedBy, "created_by user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	err = client.Channel.Create().
		SetWorkspaceID(ch.WorkspaceID).
		SetCreatedByID(createdBy).
		SetName(ch.Name).
		SetChannelType(string(ch.Type)).
		SetDmKey(key).
		OnConflictColumns(channel.FieldWorkspaceID, channel.FieldDmKey).
		DoNothing().
		Exec(ctx)
	if err := ignoreConflict(err); err != nil {
		return nil, err
	}

	c, err := client.Channel.Query().
		Where(channel.WorkspaceID(ch.WorkspaceID), channel.DmKey(key)).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return utils.ChannelToEntity(c), nil
}

func (r *channelRepository) FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
	uID, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(
			channel.WorkspaceID(workspaceID),
			channel.ChannelTypeIn("dm", "group_dm"),
			channel.HasMembersWith(channelmember.UserID(uID)),
		).
		Order(ent.Desc(channel.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return channelsToEntities(channels), nil
}

func (r *channelRepository) FindByNames(ctx context.Context, workspaceID string, names []string) ([]*entity.Channel, error) {
	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(
			channel.WorkspaceID(workspaceID),
			channel.NameIn(names...),
			channel.ChannelTypeIn(string(entity.ChannelTypePublic), string(entity.ChannelTypePrivate)),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return channelsToEntities(channels), nil
}

func (r *channelRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Channel, error) {
	parsedIDs, err := utils.ParseUUIDs(ids, "channel ID")
	if err != nil {
		return nil, err
	}
	channels, err := transaction.ResolveClient(ctx, r.client).Channel.Query().
		Where(channel.IDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return channelsToEntities(channels), nil
}

func (r *channelRepository) FindDescendants(ctx context.Context, ch *entity.Channel) ([]*entity.Channel, error) {
	client := transaction.ResolveClient(ctx, r.client)
	channels, err := client.Channel.Query().
		Where(
			channel.WorkspaceID(ch.WorkspaceID),
			channel.NameHasPrefix(ch.Name+"/"),
			channel.ChannelTypeIn(string(entity.ChannelTypePublic), string(entity.ChannelTypePrivate)),
		).
		Order(ent.Asc(channel.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return channelsToEntities(channels), nil
}

func channelsToEntities(channels []*ent.Channel) []*entity.Channel {
	result := make([]*entity.Channel, 0, len(channels))
	for _, c := range channels {
		result = append(result, utils.ChannelToEntity(c))
	}
	return result
}
