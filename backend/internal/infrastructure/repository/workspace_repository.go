package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/models"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domainrepository.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) FindByID(ctx context.Context, id string) (*entity.Workspace, error) {
	workspaceID, err := utils.ParseUUID(id, "workspace ID")
	if err != nil {
		return nil, err
	}

	var model models.Workspace
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}

func (r *workspaceRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.Workspace, error) {
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var models []models.Workspace
	if err := r.db.WithContext(ctx).
		Joins("JOIN workspace_members ON workspaces.id = workspace_members.workspace_id").
		Where("workspace_members.user_id = ?", uid).
		Order("workspaces.created_at desc").
		Find(&models).Error; err != nil {
		return nil, err
	}

	workspaces := make([]*entity.Workspace, len(models))
	for i, model := range models {
		workspaces[i] = model.ToEntity()
	}

	return workspaces, nil
}

func (r *workspaceRepository) Create(ctx context.Context, workspace *entity.Workspace) error {
	createdBy, err := utils.ParseUUID(workspace.CreatedBy, "created_by user ID")
	if err != nil {
		return err
	}

	model := &models.Workspace{}
	model.FromEntity(workspace)
	model.CreatedBy = createdBy

	if workspace.ID != "" {
		workspaceID, err := utils.ParseUUID(workspace.ID, "workspace ID")
		if err != nil {
			return err
		}
		model.ID = workspaceID
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}

	*workspace = *model.ToEntity()
	return nil
}

func (r *workspaceRepository) Update(ctx context.Context, workspace *entity.Workspace) error {
	workspaceID, err := utils.ParseUUID(workspace.ID, "workspace ID")
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"name":        workspace.Name,
		"description": workspace.Description,
		"icon_url":    workspace.IconURL,
	}

	if err := r.db.WithContext(ctx).Model(&models.Workspace{}).Where("id = ?", workspaceID).Updates(updates).Error; err != nil {
		return err
	}

	var updated models.Workspace
	if err := r.db.WithContext(ctx).Where("id = ?", workspaceID).First(&updated).Error; err != nil {
		return err
	}

	workspace.UpdatedAt = updated.UpdatedAt
	return nil
}

func (r *workspaceRepository) Delete(ctx context.Context, id string) error {
	workspaceID, err := utils.ParseUUID(id, "workspace ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.Workspace{}, "id = ?", workspaceID).Error
}

func (r *workspaceRepository) AddMember(ctx context.Context, member *entity.WorkspaceMember) error {
	workspaceID, err := utils.ParseUUID(member.WorkspaceID, "workspace ID")
	if err != nil {
		return err
	}

	userID, err := utils.ParseUUID(member.UserID, "user ID")
	if err != nil {
		return err
	}

	model := &models.WorkspaceMember{}
	model.FromEntity(member)
	model.WorkspaceID = workspaceID
	model.UserID = userID

	return r.db.WithContext(ctx).Create(model).Error
}

func (r *workspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID string, userID string, role entity.WorkspaceRole) error {
	wsID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", wsID, uid).
		Update("role", string(role)).Error
}

func (r *workspaceRepository) RemoveMember(ctx context.Context, workspaceID string, userID string) error {
	wsID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&models.WorkspaceMember{}, "workspace_id = ? AND user_id = ?", wsID, uid).Error
}

func (r *workspaceRepository) FindMembersByWorkspaceID(ctx context.Context, workspaceID string) ([]*entity.WorkspaceMember, error) {
	wsID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	var models []models.WorkspaceMember
	if err := r.db.WithContext(ctx).Where("workspace_id = ?", wsID).Order("joined_at asc").Find(&models).Error; err != nil {
		return nil, err
	}

	members := make([]*entity.WorkspaceMember, len(models))
	for i, model := range models {
		members[i] = model.ToEntity()
	}

	return members, nil
}

func (r *workspaceRepository) FindMember(ctx context.Context, workspaceID string, userID string) (*entity.WorkspaceMember, error) {
	wsID, err := utils.ParseUUID(workspaceID, "workspace ID")
	if err != nil {
		return nil, err
	}

	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return nil, err
	}

	var model models.WorkspaceMember
	if err := r.db.WithContext(ctx).Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToEntity(), nil
}
