package repository

import (
	"context"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/lib/pq"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type channelRepository struct {
	client *ent.Client
}

func NewChannelRepository(client *ent.Client) domainrepository.ChannelRepository {
	return &channelRepository{client: client}
}

func (r *channelRepository) query(ctx context.Context) *ent.ChannelQuery {
	return transaction.ResolveClient(ctx, r.client).Channel.Query()
}

func (r *channelRepository) all(ctx context.Context, query *ent.ChannelQuery) ([]*entity.Channel, error) {
	channels, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(channels, channelToEntity), nil
}

var namedChannelTypes = channel.ChannelTypeIn(string(entity.ChannelTypePublic), string(entity.ChannelTypePrivate))

func (r *channelRepository) FindByID(ctx context.Context, id string) (*entity.Channel, error) {
	channelID, err := parseUUID(id, "channel ID")
	if err != nil {
		return nil, err
	}
	c, err := orNil(r.query(ctx).Where(channel.ID(channelID)).Only(ctx))
	if c == nil {
		return nil, err
	}
	return channelToEntity(c), nil
}

func (r *channelRepository) FindBrowsableChannels(ctx context.Context, workspaceID, userID string) ([]*entity.Channel, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	return r.all(ctx, r.query(ctx).Where(viewableChannel(workspaceID, uid), namedChannelTypes).Order(ent.Asc(channel.FieldName)))
}

func (r *channelRepository) SearchBrowsableChannels(ctx context.Context, workspaceID, userID string, filter domainrepository.BrowsableChannelFilter) ([]*entity.Channel, int, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, 0, err
	}
	isMember := channel.HasMembersWith(channelmember.UserID(uid))
	query := r.query(ctx).Where(viewableChannel(workspaceID, uid), namedChannelTypes)
	if keyword := strings.TrimSpace(filter.Query); keyword != "" {
		query.Where(channel.Or(channel.NameContainsFold(keyword), channel.DescriptionContainsFold(keyword)))
	}
	switch filter.Membership {
	case domainrepository.BrowsableChannelMembershipJoined:
		query.Where(isMember)
	case domainrepository.BrowsableChannelMembershipNotJoined:
		query.Where(channel.Not(isMember))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	order := []channel.OrderOption{channel.ByName()}
	if filter.Sort == domainrepository.BrowsableChannelSortMemberCount {
		order = append([]channel.OrderOption{channel.ByMembersCount(sql.OrderDesc())}, order...)
	}
	channels, err := r.all(ctx, query.Order(order...).Limit(filter.Limit).Offset(filter.Offset))
	return channels, total, err
}

// スレッドの返信と削除済みを除いた最後のメッセージの投稿日時
const lastMessageAtSQL = `
	SELECT channel_id, MAX(created_at) FROM message
	WHERE channel_id = ANY($1::uuid[]) AND parent_id IS NULL AND deleted_at IS NULL
	GROUP BY channel_id`

func (r *channelRepository) FindLastMessageAtBatch(ctx context.Context, channelIDs []string) (map[string]time.Time, error) {
	return queryByID[time.Time](ctx, r.client, lastMessageAtSQL, pq.Array(channelIDs))
}

const memberCountSQL = `
	SELECT channel_id, COUNT(*) FROM channel_member
	WHERE channel_id = ANY($1::uuid[])
	GROUP BY channel_id`

func (r *channelRepository) CountMembersBatch(ctx context.Context, channelIDs []string) (map[string]int, error) {
	return queryByID[int](ctx, r.client, memberCountSQL, pq.Array(channelIDs))
}

func (r *channelRepository) Create(ctx context.Context, ch *entity.Channel) error {
	createdBy, err := parseUUID(ch.CreatedBy, "created_by user ID")
	if err != nil {
		return err
	}
	c, err := transaction.ResolveClient(ctx, r.client).Channel.Create().
		SetNillableID(parseUUIDPtr(&ch.ID)).
		SetWorkspaceID(ch.WorkspaceID).
		SetCreatedByID(createdBy).
		SetName(ch.Name).
		SetChannelType(string(ch.Type)).
		SetNillableDescription(ch.Description).
		SetNillableParentID(parseUUIDPtr(ch.ParentID)).
		Save(ctx)
	if err != nil {
		return err
	}
	*ch = *channelToEntity(c)
	return nil
}

func (r *channelRepository) Update(ctx context.Context, ch *entity.Channel) error {
	channelID, err := parseUUID(ch.ID, "channel ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).Channel.UpdateOneID(channelID).
		SetName(ch.Name).
		SetChannelType(string(ch.Type))
	if ch.Description != nil {
		builder.SetDescription(*ch.Description)
	} else {
		builder.ClearDescription()
	}
	return builder.Exec(ctx)
}

func (r *channelRepository) FindAccessibleChannels(ctx context.Context, workspaceID, userID string) ([]*entity.Channel, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	// DM・グループ DM は ListDirectMessages で返す
	return r.all(ctx, r.query(ctx).
		Where(channel.WorkspaceID(workspaceID), channel.HasMembersWith(channelmember.UserID(uid)), namedChannelTypes).
		Order(ent.Asc(channel.FieldName)))
}

// FindOrCreateDM は 2 人の DM を返します。なければ作ります
func (r *channelRepository) FindOrCreateDM(ctx context.Context, workspaceID string, userID1 string, userID2 string) (*entity.Channel, error) {
	return r.findOrCreateByDMKey(ctx, &entity.Channel{
		WorkspaceID: workspaceID,
		Name:        "dm_" + userID1 + "_" + userID2,
		Type:        entity.ChannelTypeDM,
		CreatedBy:   userID1,
	}, "dm:", []string{userID1, userID2})
}

// FindOrCreateGroupDM はメンバーがまったく同じグループ DM を返します
func (r *channelRepository) FindOrCreateGroupDM(ctx context.Context, workspaceID string, creatorID string, memberIDs []string) (*entity.Channel, error) {
	return r.findOrCreateByDMKey(ctx, &entity.Channel{
		WorkspaceID: workspaceID,
		Name:        "group_dm_" + creatorID,
		Type:        entity.ChannelTypeGroupDM,
		CreatedBy:   creatorID,
	}, "g:", memberIDs)
}

// findOrCreateByDMKey は参加者の組で一意な DM を返し、参加者を揃えます。同時に作られても dm_key の一意制約で 1 つにまとまる
func (r *channelRepository) findOrCreateByDMKey(ctx context.Context, ch *entity.Channel, keyPrefix string, memberIDs []string) (*entity.Channel, error) {
	createdBy, err := parseUUID(ch.CreatedBy, "created_by user ID")
	if err != nil {
		return nil, err
	}
	sorted := slices.Clone(memberIDs)
	slices.Sort(sorted)
	sorted = slices.Compact(sorted)
	key := keyPrefix + strings.Join(sorted, ",")
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
	c, err := r.query(ctx).Where(channel.WorkspaceID(ch.WorkspaceID), channel.DmKey(key)).Only(ctx)
	if err != nil {
		return nil, err
	}
	for _, id := range sorted {
		uid, err := parseUUID(id, "user ID")
		if err != nil {
			return nil, err
		}
		err = client.ChannelMember.Create().
			SetChannelID(c.ID).
			SetUserID(uid).
			SetRole(string(entity.ChannelRoleMember)).
			OnConflictColumns(channelmember.FieldChannelID, channelmember.FieldUserID).
			DoNothing().
			Exec(ctx)
		if err := ignoreConflict(err); err != nil {
			return nil, err
		}
	}
	return channelToEntity(c), nil
}

func (r *channelRepository) FindUserDMs(ctx context.Context, workspaceID string, userID string) ([]*entity.Channel, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	return r.all(ctx, r.query(ctx).
		Where(
			channel.WorkspaceID(workspaceID),
			channel.ChannelTypeIn(string(entity.ChannelTypeDM), string(entity.ChannelTypeGroupDM)),
			channel.HasMembersWith(channelmember.UserID(uid)),
		).
		Order(ent.Desc(channel.FieldUpdatedAt)))
}

func (r *channelRepository) FindByNames(ctx context.Context, workspaceID string, names []string) ([]*entity.Channel, error) {
	return r.all(ctx, r.query(ctx).Where(channel.WorkspaceID(workspaceID), channel.NameIn(names...), namedChannelTypes))
}

func (r *channelRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.Channel, error) {
	parsedIDs, err := parseUUIDs(ids, "channel ID")
	if err != nil {
		return nil, err
	}
	return r.all(ctx, r.query(ctx).Where(channel.IDIn(parsedIDs...)))
}

func (r *channelRepository) FindDescendants(ctx context.Context, parents []*entity.Channel) ([]*entity.Channel, error) {
	if len(parents) == 0 {
		return nil, nil
	}
	prefixes := make([]predicate.Channel, len(parents))
	for idx, parent := range parents {
		prefixes[idx] = channel.And(channel.WorkspaceID(parent.WorkspaceID), channel.NameHasPrefix(parent.Name+"/"))
	}
	return r.all(ctx, r.query(ctx).
		Where(channel.Or(prefixes...), namedChannelTypes).
		Order(ent.Asc(channel.FieldName)))
}

// LIKE の前方一致で使う。名前に含まれる _ をワイルドカードとして扱わない
var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

const renameDescendantsSQL = `
UPDATE channel SET name = $3 || substr(name, char_length($2) + 1), updated_at = now()
WHERE workspace_id = $1 AND name LIKE $4 AND channel_type IN ('public', 'private')`

func (r *channelRepository) RenameDescendants(ctx context.Context, workspaceID, from, to string) error {
	_, err := transaction.ResolveClient(ctx, r.client).
		ExecContext(ctx, renameDescendantsSQL, workspaceID, from, to, likeEscaper.Replace(from+"/")+"%")
	return err
}
