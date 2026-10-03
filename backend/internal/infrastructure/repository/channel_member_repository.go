package repository

import (
	"context"
	stdsql "database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/predicate"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type channelMemberRepository struct {
	client *ent.Client
}

func NewChannelMemberRepository(client *ent.Client) domainrepository.ChannelMemberRepository {
	return &channelMemberRepository{client: client}
}

func (r *channelMemberRepository) query(ctx context.Context) *ent.ChannelMemberQuery {
	return transaction.ResolveClient(ctx, r.client).ChannelMember.Query()
}

func memberOf(channelID, userID string) (predicate.ChannelMember, error) {
	cid, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	return channelmember.And(channelmember.ChannelID(cid), channelmember.UserID(uid)), nil
}

func (r *channelMemberRepository) FindMember(ctx context.Context, channelID, userID string) (*entity.ChannelMember, error) {
	where, err := memberOf(channelID, userID)
	if err != nil {
		return nil, err
	}
	cm, err := orNil(r.query(ctx).Where(where).Only(ctx))
	if cm == nil {
		return nil, err
	}
	return channelMemberToEntity(cm), nil
}

func (r *channelMemberRepository) AddMember(ctx context.Context, member *entity.ChannelMember) error {
	cid, err := parseUUID(member.ChannelID, "channel ID")
	if err != nil {
		return err
	}
	uid, err := parseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}
	err = transaction.ResolveClient(ctx, r.client).ChannelMember.Create().
		SetChannelID(cid).
		SetUserID(uid).
		SetRole(string(member.Role)).
		OnConflictColumns(channelmember.FieldChannelID, channelmember.FieldUserID).
		DoNothing().
		Exec(ctx)
	// 衝突して挿入しなかったときは RETURNING が行を返さない
	if errors.Is(err, stdsql.ErrNoRows) {
		return domerr.ErrAlreadyMember
	}
	return err
}

func (r *channelMemberRepository) RemoveMember(ctx context.Context, channelID, userID string) error {
	where, err := memberOf(channelID, userID)
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).ChannelMember.Delete().Where(where).Exec(ctx)
	return err
}

func (r *channelMemberRepository) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	where, err := memberOf(channelID, userID)
	if err != nil {
		return false, err
	}
	return r.query(ctx).Where(where).Exist(ctx)
}

func (r *channelMemberRepository) CountAdmins(ctx context.Context, channelID string) (int, error) {
	cid, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return 0, err
	}
	return r.query(ctx).Where(channelmember.ChannelID(cid), channelmember.Role(string(entity.ChannelRoleAdmin))).Count(ctx)
}

func (r *channelMemberRepository) UpdateMemberRole(ctx context.Context, channelID, userID string, role entity.ChannelRole) error {
	where, err := memberOf(channelID, userID)
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).ChannelMember.Update().Where(where).SetRole(string(role)).Exec(ctx)
}

func (r *channelMemberRepository) FindJoinedChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error) {
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	cids, err := parseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	members, err := r.query(ctx).Where(channelmember.UserID(uid), channelmember.ChannelIDIn(cids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return idSet(members, func(m *ent.ChannelMember) uuid.UUID { return m.ChannelID }), nil
}

func (r *channelMemberRepository) FindMemberIDsIn(ctx context.Context, channelID string, userIDs []string) (map[string]bool, error) {
	cid, err := parseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	uids, err := parseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}
	members, err := r.query(ctx).Where(channelmember.ChannelID(cid), channelmember.UserIDIn(uids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return idSet(members, func(m *ent.ChannelMember) uuid.UUID { return m.UserID }), nil
}

func (r *channelMemberRepository) FindMembersByChannelIDs(ctx context.Context, channelIDs []string) ([]*entity.ChannelMember, error) {
	cids, err := parseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	members, err := r.query(ctx).
		Where(channelmember.ChannelIDIn(cids...)).
		Order(ent.Asc(channelmember.FieldJoinedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(members, channelMemberToEntity), nil
}
