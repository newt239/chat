package repository

import (
	"context"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/channel"
	"github.com/newt239/chat/ent/channelmember"
	"github.com/newt239/chat/ent/user"
	"github.com/newt239/chat/internal/domain/entity"
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

func (r *channelMemberRepository) FindByChannelAndUser(ctx context.Context, channelID, userID string) (*entity.ChannelMember, error) {
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
			channelmember.HasChannelWith(channel.ID(cid)),
			channelmember.HasUserWith(user.ID(uid)),
		).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
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

	_, err = client.ChannelMember.Create().
		SetChannelID(cid).
		SetUserID(uid).
		SetRole(string(member.Role)).
		Save(ctx)
	if err != nil {
		return err
	}

	// Load edges
	cm, err := client.ChannelMember.Query().
		Where(
			channelmember.HasChannelWith(channel.ID(cid)),
			channelmember.HasUserWith(user.ID(uid)),
		).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
		Only(ctx)
	if err != nil {
		return err
	}

	*member = *utils.ChannelMemberToEntity(cm)
	return nil
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
			channelmember.HasChannelWith(channel.ID(cid)),
			channelmember.HasUserWith(user.ID(uid)),
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
		Where(channelmember.HasChannelWith(channel.ID(cid))).
		WithChannel(func(q *ent.ChannelQuery) {
			q.WithWorkspace().WithCreatedBy()
		}).
		WithUser().
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
			channelmember.HasChannelWith(channel.ID(cid)),
			channelmember.HasUserWith(user.ID(uid)),
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
			channelmember.HasChannelWith(channel.ID(cid)),
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
			channelmember.HasChannelWith(channel.ID(cid)),
			channelmember.HasUserWith(user.ID(uid)),
		).
		SetRole(string(role)).
		Save(ctx)

	return err
}
