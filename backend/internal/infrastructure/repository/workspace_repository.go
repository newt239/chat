package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/example/chat/internal/domain"
	"github.com/example/chat/internal/infrastructure/db"
)

type workspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) domain.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

func (r *workspaceRepository) FindByID(id string) (*domain.Workspace, error) {
	workspaceID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	var dbWorkspace db.Workspace
	if err := r.db.Where("id = ?", workspaceID).First(&dbWorkspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toWorkspaceDomain(&dbWorkspace), nil
}

func (r *workspaceRepository) FindByUserID(userID string) ([]*domain.Workspace, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbWorkspaces []db.Workspace
	if err := r.db.
		Joins("JOIN workspace_members ON workspaces.id = workspace_members.workspace_id").
		Where("workspace_members.user_id = ?", uid).
		Order("workspaces.created_at desc").
		Find(&dbWorkspaces).Error; err != nil {
		return nil, err
	}

	workspaces := make([]*domain.Workspace, len(dbWorkspaces))
	for i, w := range dbWorkspaces {
		workspaces[i] = toWorkspaceDomain(&w)
	}

	return workspaces, nil
}

func (r *workspaceRepository) Create(workspace *domain.Workspace) error {
	createdBy, err := uuid.Parse(workspace.CreatedBy)
	if err != nil {
		return errors.New("invalid created_by user ID format")
	}

	dbWorkspace := &db.Workspace{
		Name:        workspace.Name,
		Description: workspace.Description,
		IconURL:     workspace.IconURL,
		CreatedBy:   createdBy,
	}

	if workspace.ID != "" {
		workspaceID, err := uuid.Parse(workspace.ID)
		if err != nil {
			return errors.New("invalid workspace ID format")
		}
		dbWorkspace.ID = workspaceID
	}

	if err := r.db.Create(dbWorkspace).Error; err != nil {
		return err
	}

	workspace.ID = dbWorkspace.ID.String()
	workspace.CreatedAt = dbWorkspace.CreatedAt
	workspace.UpdatedAt = dbWorkspace.UpdatedAt

	return nil
}

func (r *workspaceRepository) Update(workspace *domain.Workspace) error {
	workspaceID, err := uuid.Parse(workspace.ID)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	updates := map[string]interface{}{
		"name":        workspace.Name,
		"description": workspace.Description,
		"icon_url":    workspace.IconURL,
	}

	if err := r.db.Model(&db.Workspace{}).Where("id = ?", workspaceID).Updates(updates).Error; err != nil {
		return err
	}

	// Fetch updated record
	var updated db.Workspace
	if err := r.db.Where("id = ?", workspaceID).First(&updated).Error; err != nil {
		return err
	}

	workspace.UpdatedAt = updated.UpdatedAt

	return nil
}

func (r *workspaceRepository) Delete(id string) error {
	workspaceID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	return r.db.Delete(&db.Workspace{}, "id = ?", workspaceID).Error
}

func (r *workspaceRepository) AddMember(member *domain.WorkspaceMember) error {
	workspaceID, err := uuid.Parse(member.WorkspaceID)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	userID, err := uuid.Parse(member.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	dbMember := &db.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        string(member.Role),
	}

	return r.db.Create(dbMember).Error
}

func (r *workspaceRepository) UpdateMemberRole(workspaceID, userID string, role domain.WorkspaceRole) error {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return r.db.Model(&db.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", wsID, uid).
		Update("role", string(role)).Error
}

func (r *workspaceRepository) RemoveMember(workspaceID, userID string) error {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return errors.New("invalid workspace ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	return r.db.Delete(&db.WorkspaceMember{}, "workspace_id = ? AND user_id = ?", wsID, uid).Error
}

func (r *workspaceRepository) FindMembersByWorkspaceID(workspaceID string) ([]*domain.WorkspaceMember, error) {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	var dbMembers []db.WorkspaceMember
	if err := r.db.Where("workspace_id = ?", wsID).Order("joined_at asc").Find(&dbMembers).Error; err != nil {
		return nil, err
	}

	members := make([]*domain.WorkspaceMember, len(dbMembers))
	for i, m := range dbMembers {
		members[i] = toWorkspaceMemberDomain(&m)
	}

	return members, nil
}

func (r *workspaceRepository) FindMember(workspaceID, userID string) (*domain.WorkspaceMember, error) {
	wsID, err := uuid.Parse(workspaceID)
	if err != nil {
		return nil, errors.New("invalid workspace ID format")
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	var dbMember db.WorkspaceMember
	if err := r.db.Where("workspace_id = ? AND user_id = ?", wsID, uid).First(&dbMember).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toWorkspaceMemberDomain(&dbMember), nil
}

func toWorkspaceDomain(dbWorkspace *db.Workspace) *domain.Workspace {
	return &domain.Workspace{
		ID:          dbWorkspace.ID.String(),
		Name:        dbWorkspace.Name,
		Description: dbWorkspace.Description,
		IconURL:     dbWorkspace.IconURL,
		CreatedBy:   dbWorkspace.CreatedBy.String(),
		CreatedAt:   dbWorkspace.CreatedAt,
		UpdatedAt:   dbWorkspace.UpdatedAt,
	}
}

func toWorkspaceMemberDomain(dbMember *db.WorkspaceMember) *domain.WorkspaceMember {
	return &domain.WorkspaceMember{
		WorkspaceID: dbMember.WorkspaceID.String(),
		UserID:      dbMember.UserID.String(),
		Role:        domain.WorkspaceRole(dbMember.Role),
		JoinedAt:    dbMember.JoinedAt,
	}
}
