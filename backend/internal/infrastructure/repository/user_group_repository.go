package repository

import (
	"context"
	stdsql "database/sql"
	"errors"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/usergroup"
	"github.com/newt239/chat/ent/usergroupmember"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
)

type userGroupRepository struct {
	client *ent.Client
}

func NewUserGroupRepository(client *ent.Client) domainrepository.UserGroupRepository {
	return &userGroupRepository{client: client}
}

func (r *userGroupRepository) query(ctx context.Context) *ent.UserGroupQuery {
	return transaction.ResolveClient(ctx, r.client).UserGroup.Query()
}

func (r *userGroupRepository) FindByID(ctx context.Context, id string) (*entity.UserGroup, error) {
	gid, err := parseUUID(id, "group ID")
	if err != nil {
		return nil, err
	}
	ug, err := orNil(r.query(ctx).Where(usergroup.ID(gid)).Only(ctx))
	if ug == nil {
		return nil, err
	}
	return userGroupToEntity(ug), nil
}

func (r *userGroupRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.UserGroup, error) {
	parsedIDs, err := parseUUIDs(ids, "group ID")
	if err != nil {
		return nil, err
	}
	groups, err := r.query(ctx).Where(usergroup.IDIn(parsedIDs...)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(groups, userGroupToEntity), nil
}

func (r *userGroupRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.UserGroup, error) {
	groups, err := r.query(ctx).Where(usergroup.WorkspaceID(workspaceID)).All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(groups, userGroupToEntity), nil
}

func (r *userGroupRepository) FindByName(ctx context.Context, workspaceID string, name string) (*entity.UserGroup, error) {
	ug, err := orNil(r.query(ctx).Where(usergroup.WorkspaceID(workspaceID), usergroup.Name(name)).Only(ctx))
	if ug == nil {
		return nil, err
	}
	return userGroupToEntity(ug), nil
}

func (r *userGroupRepository) Create(ctx context.Context, group *entity.UserGroup) error {
	createdBy, err := parseUUID(group.CreatedBy, "created by user ID")
	if err != nil {
		return err
	}
	ug, err := transaction.ResolveClient(ctx, r.client).UserGroup.Create().
		SetNillableID(parseUUIDPtr(&group.ID)).
		SetWorkspaceID(group.WorkspaceID).
		SetCreatedByID(createdBy).
		SetName(group.Name).
		SetNillableDescription(group.Description).
		Save(ctx)
	if err != nil {
		return err
	}
	*group = *userGroupToEntity(ug)
	return nil
}

func (r *userGroupRepository) Update(ctx context.Context, group *entity.UserGroup) error {
	gid, err := parseUUID(group.ID, "group ID")
	if err != nil {
		return err
	}
	builder := transaction.ResolveClient(ctx, r.client).UserGroup.UpdateOneID(gid).SetName(group.Name)
	if group.Description != nil {
		builder.SetDescription(*group.Description)
	} else {
		builder.ClearDescription()
	}
	ug, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	*group = *userGroupToEntity(ug)
	return nil
}

func (r *userGroupRepository) Delete(ctx context.Context, id string) error {
	gid, err := parseUUID(id, "group ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).UserGroup.DeleteOneID(gid).Exec(ctx)
}

func (r *userGroupRepository) AddMember(ctx context.Context, member *entity.UserGroupMember) error {
	gid, err := parseUUID(member.GroupID, "group ID")
	if err != nil {
		return err
	}
	uid, err := parseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}
	err = transaction.ResolveClient(ctx, r.client).UserGroupMember.Create().
		SetGroupID(gid).
		SetUserID(uid).
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
	gid, err := parseUUID(groupID, "group ID")
	if err != nil {
		return err
	}
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	_, err = transaction.ResolveClient(ctx, r.client).UserGroupMember.Delete().
		Where(usergroupmember.GroupID(gid), usergroupmember.UserID(uid)).
		Exec(ctx)
	return err
}

func (r *userGroupRepository) FindMembersByGroupIDs(ctx context.Context, groupIDs []string) ([]*entity.UserGroupMember, error) {
	gids, err := parseUUIDs(groupIDs, "group ID")
	if err != nil {
		return nil, err
	}
	members, err := transaction.ResolveClient(ctx, r.client).UserGroupMember.Query().
		Where(usergroupmember.GroupIDIn(gids...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return convertAll(members, func(m *ent.UserGroupMember) *entity.UserGroupMember {
		return &entity.UserGroupMember{GroupID: m.GroupID.String(), UserID: m.UserID.String()}
	}), nil
}

func (r *userGroupRepository) IsMember(ctx context.Context, groupID string, userID string) (bool, error) {
	gid, err := parseUUID(groupID, "group ID")
	if err != nil {
		return false, err
	}
	uid, err := parseUUID(userID, "user ID")
	if err != nil {
		return false, err
	}
	return transaction.ResolveClient(ctx, r.client).UserGroupMember.Query().
		Where(usergroupmember.GroupID(gid), usergroupmember.UserID(uid)).
		Exist(ctx)
}
