package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/models"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type userGroupRepository struct {
	db *gorm.DB
}

func NewUserGroupRepository(db *gorm.DB) domainrepository.UserGroupRepository {
	return &userGroupRepository{db: db}
}

func (r *userGroupRepository) FindByID(ctx context.Context, id string) (*entity.UserGroup, error) {
	groupID, err := utils.ParseUUID(id, "group ID")
	if err != nil {
		return nil, err
	}

	var model models.UserGroup
	if err := r.db.WithContext(ctx).Where("id = ?", groupID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *userGroupRepository) FindByIDs(ctx context.Context, ids []string) ([]*entity.UserGroup, error) {
	if len(ids) == 0 {
		return []*entity.UserGroup{}, nil
	}

	groupIDs := make([]interface{}, len(ids))
	for i, id := range ids {
		groupID, err := utils.ParseUUID(id, "group ID")
		if err != nil {
			return nil, err
		}
		groupIDs[i] = groupID
	}

	var models []models.UserGroup
	if err := r.db.WithContext(ctx).Where("id IN ?", groupIDs).Find(&models).Error; err != nil {
		return nil, err
	}

	groups := make([]*entity.UserGroup, len(models))
	for i, model := range models {
		groups[i] = model.ToEntity()
	}

	return groups, nil
}

func (r *userGroupRepository) FindByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.UserGroup, error) {
	wsID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	var models []models.UserGroup
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", wsID).Order("name asc").Find(&models).Error; err != nil {
		return nil, err
	}

	groups := make([]*entity.UserGroup, len(models))
	for i, model := range models {
		groups[i] = model.ToEntity()
	}

	return groups, nil
}

func (r *userGroupRepository) FindByName(ctx context.Context, workspaceID string, name string) (*entity.UserGroup, error) {
	wsID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	var model models.UserGroup
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND name = ?", wsID, name).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *userGroupRepository) Create(ctx context.Context, group *entity.UserGroup) error {
	workspaceID, err := utils.ParseUUID(group.WorkspaceID, "workspace ID")
	if err != nil {
		return err
	}

	createdBy, err := utils.ParseUUID(group.CreatedBy, "created by ID")
	if err != nil {
		return err
	}

	model := &models.UserGroup{}
	model.FromEntity(group)
	model.WorkspaceID = workspaceID
	model.CreatedBy = createdBy

	if group.ID != "" {
		groupID, err := utils.ParseUUID(group.ID, "group ID")
		if err != nil {
			return err
		}
		model.ID = groupID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*group = *model.ToEntity()
	return nil
}

func (r *userGroupRepository) Update(ctx context.Context, group *entity.UserGroup) error {
	groupID, err := utils.ParseUUID(group.ID, "group ID")
	if err != nil {
		return err
	}

	now := time.Now()
	updates := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"updated_at":  now,
	}

	if err := r.db.WithContext(ctx).Model(&models.UserGroup{}).Where("id = ?", groupID).Updates(updates).Error; err != nil {
		return err
	}

	group.UpdatedAt = now

	return nil
}

func (r *userGroupRepository) Delete(ctx context.Context, id string) error {
	groupID, err := utils.ParseUUID(id, "group ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.UserGroup{}, "id = ?", groupID).Error
}

func (r *userGroupRepository) AddMember(ctx context.Context, member *entity.UserGroupMember) error {
	groupID, err := utils.ParseUUID(member.GroupID, "group ID")
	if err != nil {
		return err
	}

	userID, err := utils.ParseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &models.UserGroupMember{}
	model.FromEntity(member)
	model.GroupID = groupID
	model.UserID = userID

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *userGroupRepository) RemoveMember(ctx context.Context, groupID string, userID string) error {
	gID, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.UserGroupMember{}, "group_id = ? AND user_id = ?", gID, uid).Error
}

func (r *userGroupRepository) FindMembersByGroupID(ctx context.Context, groupID string) ([]*entity.UserGroupMember, error) {
	gID, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return nil, err
	}

	var models []models.UserGroupMember
	if err := r.db.WithContext(ctx).Where("group_id = ?", gID).Order("joined_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	members := make([]*entity.UserGroupMember, len(models))
	for i, model := range models {
		members[i] = model.ToEntity()
	}

	return members, nil
}

func (r *userGroupRepository) FindGroupsByUserID(ctx context.Context, userID string) ([]*entity.UserGroup, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var models []models.UserGroup
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_group_members ON user_groups.id = user_group_members.group_id").
		Where("user_group_members.user_id = ?", uid).
		Order("user_groups.name asc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	groups := make([]*entity.UserGroup, len(models))
	for i, model := range models {
		groups[i] = model.ToEntity()
	}

	return groups, nil
}

func (r *userGroupRepository) IsMember(ctx context.Context, groupID string, userID string) (bool, error) {
	gID, err := utils.ParseUUID(groupID, "group ID")
	if err != nil {
		return false, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return false, err
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&models.UserGroupMember{}).
		Where("group_id = ? AND user_id = ?", gID, uid).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
