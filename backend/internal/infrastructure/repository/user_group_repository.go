package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type userGroupRepository struct {
	db *gorm.DB
}

func NewUserGroupRepository(db *gorm.DB) domain.UserGroupRepository {
	return &userGroupRepository{db: db}
}

func (r *userGroupRepository) FindByID(id string) (*domain.UserGroup, error) {
	groupID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}

	var dbGroup db.UserGroup
	if err := r.db.Where("id = ?", groupID).First(&dbGroup).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserGroupDomain(&dbGroup), nil
}

func (r *userGroupRepository) FindByWorkspaceID(workspaceID string) ([]*domain.UserGroup, error) {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	var dbGroups []db.UserGroup
	if err := r.db.Where("workspace_id = ?", wsID).Order("name asc").Find(&dbGroups).Error; err != nil {
		return nil, err
	}

	groups := make([]*domain.UserGroup, len(dbGroups))
	for i, g := range dbGroups {
		groups[i] = toUserGroupDomain(&g)
	}

	return groups, nil
}

func (r *userGroupRepository) FindByName(workspaceID, name string) (*domain.UserGroup, error) {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	var dbGroup db.UserGroup
	if err := r.db.Where("workspace_id = ? AND name = ?", wsID, name).First(&dbGroup).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toUserGroupDomain(&dbGroup), nil
}

func (r *userGroupRepository) Create(group *domain.UserGroup) error {
	workspaceID, err := uuid.Parse(group.WorkspaceID)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	createdBy, err := uuid.Parse(group.CreatedBy)
	if err != nil {
		return errors.New("invalid created by ID format")
	}

	dbGroup := &db.UserGroup{
		WorkspaceID: workspaceID,
		Name:        group.Name,
		Description: group.Description,
		CreatedBy:   createdBy,
	}

	if group.ID != "" {
		groupID, err := uuid.Parse(group.ID)
		if err != nil {
			return errors.New("invalid group ID format")
		}
		dbGroup.ID = groupID
	}

	if err := r.db.Create(dbGroup).Error; err != nil {
		return err
	}

	group.ID = dbGroup.ID.String()
	group.CreatedAt = dbGroup.CreatedAt
	group.UpdatedAt = dbGroup.UpdatedAt

	return nil
}

func (r *userGroupRepository) Update(group *domain.UserGroup) error {
	groupID, err := uuid.Parse(group.ID)
	if err != nil {
		return errors.New("invalid group ID format")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"name":        group.Name,
		"description": group.Description,
		"updated_at":  now,
	}

	if err := r.db.Model(&db.UserGroup{}).Where("id = ?", groupID).Updates(updates).Error; err != nil {
		return err
	}

	group.UpdatedAt = now

	return nil
}

func (r *userGroupRepository) Delete(id string) error {
	groupID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid group ID format")
	}

	return r.db.Delete(&db.UserGroup{}, "id = ?", groupID).Error
}

func (r *userGroupRepository) AddMember(member *domain.UserGroupMember) error {
	groupID, err := uuid.Parse(member.GroupID)
	if err != nil {
		return errors.New("invalid group ID format")
	}

	userID, err := uuid.Parse(member.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbMember := &db.UserGroupMember{
		GroupID:  groupID,
		UserID:   userID,
		JoinedAt: member.JoinedAt,
	}

	return r.db.Create(dbMember).Error
}

func (r *userGroupRepository) RemoveMember(groupID, userID string) error {
	gID, err := uuid.Parse(groupID)
	if err != nil {
		return errors.New("invalid group ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return r.db.Delete(&db.UserGroupMember{}, "group_id = ? AND user_id = ?", gID, uid).Error
}

func (r *userGroupRepository) FindMembersByGroupID(groupID string) ([]*domain.UserGroupMember, error) {
	gID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, errors.New("invalid group ID format")
	}

	var dbMembers []db.UserGroupMember
	if err := r.db.Where("group_id = ?", gID).Order("joined_at asc").Find(&dbMembers).Error; err != nil {
		return nil, err
	}

	members := make([]*domain.UserGroupMember, len(dbMembers))
	for i, m := range dbMembers {
		members[i] = toUserGroupMemberDomain(&m)
	}

	return members, nil
}

func (r *userGroupRepository) FindGroupsByUserID(userID string) ([]*domain.UserGroup, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbGroups []db.UserGroup
	if err := r.db.
		Joins("JOIN user_group_members ON user_groups.id = user_group_members.group_id").
		Where("user_group_members.user_id = ?", uid).
		Order("user_groups.name asc").
		Find(&dbGroups).Error; err != nil {
		return nil, err
	}

	groups := make([]*domain.UserGroup, len(dbGroups))
	for i, g := range dbGroups {
		groups[i] = toUserGroupDomain(&g)
	}

	return groups, nil
}

func (r *userGroupRepository) IsMember(groupID, userID string) (bool, error) {
	gID, err := uuid.Parse(groupID)
	if err != nil {
		return false, errors.New("invalid group ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.New("invalid user ID format")
	}

	var count int64
	if err := r.db.Model(&db.UserGroupMember{}).
		Where("group_id = ? AND user_id = ?", gID, uid).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func toUserGroupDomain(dbGroup *db.UserGroup) *domain.UserGroup {
	return &domain.UserGroup{
		ID:          dbGroup.ID.String(),
		WorkspaceID: dbGroup.WorkspaceID.String(),
		Name:        dbGroup.Name,
		Description: dbGroup.Description,
		CreatedBy:   dbGroup.CreatedBy.String(),
		CreatedAt:   dbGroup.CreatedAt,
		UpdatedAt:   dbGroup.UpdatedAt,
	}
}

func toUserGroupMemberDomain(dbMember *db.UserGroupMember) *domain.UserGroupMember {
	return &domain.UserGroupMember{
		GroupID:  dbMember.GroupID.String(),
		UserID:   dbMember.UserID.String(),
		JoinedAt: dbMember.JoinedAt,
	}
}
