package repository

import (
	"context"
	stdsql "database/sql"
	"errors"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type channelMemberRepository struct {
	client *ent.Client
}

func NewChannelMemberRepository(client *ent.Client) domainrepository.ChannelMemberRepository {
	return &channelMemberRepository{client: client}
}

func (r *channelMemberRepository) FindMember(ctx context.Context, channelID, userID string) (*entity.ChannelMember, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	cm, err := client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(cid),
			channelmember.UserID(uid),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.ChannelMemberToEntity(cm), nil
}

func (r *channelMemberRepository) AddMember(ctx context.Context, member *entity.ChannelMember) error {
	cid, err := utils.ParseUUID(member.ChannelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	if member.JoinedAt.IsZero() {
		member.JoinedAt = time.Now()
	}
	err = client.ChannelMember.Create().
		SetChannelID(cid).
		SetUserID(uid).
		SetRole(string(member.Role)).
		SetJoinedAt(member.JoinedAt).
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
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.ChannelMember.Delete().
		Where(
			channelmember.ChannelID(cid),
			channelmember.UserID(uid),
		).
		Exec(ctx)

	return err
}

func (r *channelMemberRepository) FindMembers(ctx context.Context, channelID string) ([]*entity.ChannelMember, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	members, err := client.ChannelMember.Query().
		Where(channelmember.ChannelID(cid)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.ChannelMember, 0, len(members))
	for _, cm := range members {
		result = append(result, utils.ChannelMemberToEntity(cm))
	}

	return result, nil
}

func (r *channelMemberRepository) IsMember(ctx context.Context, channelID, userID string) (bool, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return false, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return false, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	exists, err := client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(cid),
			channelmember.UserID(uid),
		).
		Exist(ctx)

	return exists, err
}

func (r *channelMemberRepository) CountAdmins(ctx context.Context, channelID string) (int, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return 0, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	count, err := client.ChannelMember.Query().
		Where(
			channelmember.ChannelID(cid),
			channelmember.RoleEQ(string(entity.ChannelRoleAdmin)),
		).
		Count(ctx)

	return count, err
}

func (r *channelMemberRepository) UpdateMemberRole(ctx context.Context, channelID, userID string, role entity.ChannelRole) error {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.ChannelMember.Update().
		Where(
			channelmember.ChannelID(cid),
			channelmember.UserID(uid),
		).
		SetRole(string(role)).
		Save(ctx)

	return err
}

func (r *channelMemberRepository) FindJoinedChannelIDs(ctx context.Context, userID string, channelIDs []string) (map[string]bool, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}
	cids, err := utils.ParseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	members, err := transaction.ResolveClient(ctx, r.client).ChannelMember.Query().
		Where(channelmember.UserID(uid), channelmember.ChannelIDIn(cids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(members))
	for _, m := range members {
		result[m.ChannelID.String()] = true
	}
	return result, nil
}

func (r *channelMemberRepository) FindMemberIDsIn(ctx context.Context, channelID string, userIDs []string) (map[string]bool, error) {
	cid, err := utils.ParseUUID(channelID, "channel ID")
	if err != nil {
		return nil, err
	}
	uids, err := utils.ParseUUIDs(userIDs, "user ID")
	if err != nil {
		return nil, err
	}
	members, err := transaction.ResolveClient(ctx, r.client).ChannelMember.Query().
		Where(channelmember.ChannelID(cid), channelmember.UserIDIn(uids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(members))
	for _, m := range members {
		result[m.UserID.String()] = true
	}
	return result, nil
}

func (r *channelMemberRepository) FindMembersByChannelIDs(ctx context.Context, channelIDs []string) ([]*entity.ChannelMember, error) {
	cids, err := utils.ParseUUIDs(channelIDs, "channel ID")
	if err != nil {
		return nil, err
	}
	members, err := transaction.ResolveClient(ctx, r.client).ChannelMember.Query().
		Where(channelmember.ChannelIDIn(cids...)).
		Order(ent.Asc(channelmember.FieldJoinedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.ChannelMember, 0, len(members))
	for _, cm := range members {
		result = append(result, utils.ChannelMemberToEntity(cm))
	}
	return result, nil
}
