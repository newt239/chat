package repository

import (
	"context"
	stdsql "database/sql"
	"errors"

	domerr "github.com/newt239/chat/internal/domain/errors"

	"github.com/google/uuid"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/usergroup"
	"github.com/newt239/chat/ent/usergroupmember"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type userGroupRepository struct {
	client *ent.Client
}

func NewUserGroupRepository(client *ent.Client) domainrepository.UserGroupRepository {
	return &userGroupRepository{client: client}
}

func (r *userGroupRepository) FindByID(ctx context.Context, id string) (*entity.UserGroup, error) {
	gid, err := utils.ParseUUID(id, "group ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	ug, err := client.UserGroup.Query().
		Where(usergroup.ID(gid)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.UserGroupToEntity(ug), nil
}

func (r *userGroupRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.UserGroup, error) {
	if len(ids) == 0 {
		return []*entity.UserGroup{}, nil
	}

	// Parse all group IDs
	parsedIDs := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		parsedID, err := utils.ParseUUID(id, "group ID")
		if err != nil {
			return nil, err
		}
		parsedIDs = append(parsedIDs, parsedID)
	}

	client := transaction.ResolveClient(ctx, r.client)
	groups, err := client.UserGroup.Query().
		Where(usergroup.IDIn(parsedIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.UserGroup, 0, len(groups))
	for _, ug := range groups {
		result = append(result, utils.UserGroupToEntity(ug))
	}

	return result, nil
}

func (r *userGroupRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.UserGroup, error) {
	client := transaction.ResolveClient(ctx, r.client)
	groups, err := client.UserGroup.Query().
		Where(usergroup.WorkspaceID(workspaceID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.UserGroup, 0, len(groups))
	for _, ug := range groups {
		result = append(result, utils.UserGroupToEntity(ug))
	}

	return result, nil
}

func (r *userGroupRepository) FindByName(ctx context.Context, workspaceID string, name string) (*entity.UserGroup, error) {
	client := transaction.ResolveClient(ctx, r.client)
	ug, err := client.UserGroup.Query().
		Where(
			usergroup.WorkspaceID(workspaceID),
			usergroup.Name(name),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return utils.UserGroupToEntity(ug), nil
}

func (r *userGroupRepository) Create(ctx context.Context, group *entity.UserGroup) error {
	// workspaceID is slug (string)
	cid, err := utils.ParseUUID(group.CreatedBy, "created by user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.UserGroup.Create().
		SetWorkspaceID(group.WorkspaceID).
		SetCreatedByID(cid).
		SetName(group.Name)

	if group.ID != "" {
		gid, err := utils.ParseUUID(group.ID, "group ID")
		if err != nil {
			return err
		}
		builder = builder.SetID(gid)
	}

	if group.Description != nil {
		builder = builder.SetDescription(*group.Description)
	}

	ug, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	*group = *utils.UserGroupToEntity(ug)
	return nil
}

func (r *userGroupRepository) Update(ctx context.Context, group *entity.UserGroup) error {
	gid, err := utils.ParseUUID(group.ID, "group ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	builder := client.UserGroup.UpdateOneID(gid).
		SetName(group.Name)

	if group.Description != nil {
		builder = builder.SetDescription(*group.Description)
	}

	ug, err := builder.Save(ctx)
	if err != nil {
		return err
	}

	*group = *utils.UserGroupToEntity(ug)
	return nil
}

func (r *userGroupRepository) Delete(ctx context.Context, id string) error {
	gid, err := utils.ParseUUID(id, "group ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	return client.UserGroup.DeleteOneID(gid).Exec(ctx)
}

func (r *userGroupRepository) AddMember(ctx context.Context, member *entity.UserGroupMember) error {
	gid, err := utils.ParseUUID(member.GroupID, "group ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)

	err = client.UserGroupMember.Create().
		SetGroupID(gid).
		SetUserID(uid).
		SetJoinedAt(member.JoinedAt).
		OnConflictColumns(usergroupmember.FieldGroupID, usergroupmember.FieldUserID).
		DoNothing().
		Exec(ctx)
	// 衝突して挿入しなかったときは RETURNING が行を返さない
	if errors.Is(err, stdsql.ErrNoRows) {
		return domerr.ErrAlreadyMember
	}
	return err
}

func (r *userGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	gid, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	client := transaction.ResolveClient(ctx, r.client)
	_, err = client.UserGroupMember.Delete().
		Where(
			usergroupmember.GroupID(gid),
			usergroupmember.UserID(uid),
		).
		Exec(ctx)

	return err
}

func (r *userGroupRepository) FindMembersByGroupID(ctx context.Context, groupID string) ([]*entity.UserGroupMember, error) {
	gid, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	members, err := client.UserGroupMember.Query().
		Where(usergroupmember.GroupID(gid)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.UserGroupMember, 0, len(members))
	for _, ugm := range members {
		result = append(result, utils.UserGroupMemberToEntity(ugm))
	}

	return result, nil
}

func (r *userGroupRepository) FindGroupsByUserID(ctx context.Context, userID string) ([]*entity.UserGroup, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	groups, err := client.UserGroup.Query().
		Where(usergroup.HasMembersWith(usergroupmember.UserID(uid))).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.UserGroup, 0, len(groups))
	for _, ug := range groups {
		result = append(result, utils.UserGroupToEntity(ug))
	}

	return result, nil
}

func (r *userGroupRepository) IsMember(ctx context.Context, groupID string, userID string) (bool, error) {
	gid, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return false, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return false, err
	}

	client := transaction.ResolveClient(ctx, r.client)
	exists, err := client.UserGroupMember.Query().
		Where(
			usergroupmember.GroupID(gid),
			usergroupmember.UserID(uid),
		).
		Exist(ctx)
	if err != nil {
		return false, err
	}

	return exists, nil
}
