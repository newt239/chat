package workspace

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain"
)

var (
	ErrWorkspaceNotFound    = errors.New("workspace not found")
	ErrUnauthorized         = errors.New("unauthorized to perform this action")
	ErrInvalidRole          = errors.New("invalid workspace role")
	ErrCannotRemoveOwner    = errors.New("cannot remove workspace owner")
	ErrCannotChangeOwnerRole = errors.New("cannot change owner role")
)

// WorkspaceUseCase defines the interface for workspace operations
type WorkspaceUseCase interface {
	GetWorkspacesByUserID(userID string) (*GetWorkspacesOutput, error)
	GetWorkspace(input GetWorkspaceInput) (*GetWorkspaceOutput, error)
	CreateWorkspace(input CreateWorkspaceInput) (*CreateWorkspaceOutput, error)
	UpdateWorkspace(input UpdateWorkspaceInput) (*UpdateWorkspaceOutput, error)
	DeleteWorkspace(input DeleteWorkspaceInput) (*DeleteWorkspaceOutput, error)
	ListMembers(input ListMembersInput) (*ListMembersOutput, error)
	AddMember(input AddMemberInput) (*MemberActionOutput, error)
	UpdateMemberRole(input UpdateMemberRoleInput) (*MemberActionOutput, error)
	RemoveMember(input RemoveMemberInput) (*MemberActionOutput, error)
}

type workspaceInteractor struct {
	workspaceRepo domain.WorkspaceRepository
	userRepo      domain.UserRepository
}

// NewWorkspaceInteractor creates a new workspace use case interactor
func NewWorkspaceInteractor(
	workspaceRepo domain.WorkspaceRepository,
	userRepo domain.UserRepository,
) WorkspaceUseCase {
	return &workspaceInteractor{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (i *workspaceInteractor) GetWorkspacesByUserID(userID string) (*GetWorkspacesOutput, error) {
	// Get workspaces for the user
	workspaces, err := i.workspaceRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	// Convert to output format
	output := &GetWorkspacesOutput{
		Workspaces: make([]WorkspaceOutput, 0, len(workspaces)),
	}

	for _, ws := range workspaces {
		// Get member info to retrieve role
		member, err := i.workspaceRepo.FindMember(ws.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member info: %w", err)
		}

		output.Workspaces = append(output.Workspaces, WorkspaceOutput{
			ID:          ws.ID,
			Name:        ws.Name,
			Description: ws.Description,
			IconURL:     ws.IconURL,
			Role:        string(member.Role),
			CreatedBy:   ws.CreatedBy,
			CreatedAt:   ws.CreatedAt,
			UpdatedAt:   ws.UpdatedAt,
		})
	}

	return output, nil
}

func (i *workspaceInteractor) GetWorkspace(input GetWorkspaceInput) (*GetWorkspaceOutput, error) {
	// Check if user is a member
	member, err := i.workspaceRepo.FindMember(input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// Get workspace
	ws, err := i.workspaceRepo.FindByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	return &GetWorkspaceOutput{
		Workspace: WorkspaceOutput{
			ID:          ws.ID,
			Name:        ws.Name,
			Description: ws.Description,
			IconURL:     ws.IconURL,
			Role:        string(member.Role),
			CreatedBy:   ws.CreatedBy,
			CreatedAt:   ws.CreatedAt,
			UpdatedAt:   ws.UpdatedAt,
		},
	}, nil
}

func (i *workspaceInteractor) CreateWorkspace(input CreateWorkspaceInput) (*CreateWorkspaceOutput, error) {
	// Create workspace
	workspace := &domain.Workspace{
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.workspaceRepo.Create(workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// Add creator as owner
	member := &domain.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      input.CreatedBy,
		Role:        domain.WorkspaceRoleOwner,
		JoinedAt:    time.Now(),
	}

	if err := i.workspaceRepo.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add creator as owner: %w", err)
	}

	return &CreateWorkspaceOutput{
		Workspace: WorkspaceOutput{
			ID:          workspace.ID,
			Name:        workspace.Name,
			Description: workspace.Description,
			IconURL:     workspace.IconURL,
			Role:        string(domain.WorkspaceRoleOwner),
			CreatedBy:   workspace.CreatedBy,
			CreatedAt:   workspace.CreatedAt,
			UpdatedAt:   workspace.UpdatedAt,
		},
	}, nil
}

func (i *workspaceInteractor) UpdateWorkspace(input UpdateWorkspaceInput) (*UpdateWorkspaceOutput, error) {
	// Check if user is admin or owner
	member, err := i.workspaceRepo.FindMember(input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil || (member.Role != domain.WorkspaceRoleOwner && member.Role != domain.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	// Get existing workspace
	ws, err := i.workspaceRepo.FindByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	// Update fields
	if input.Name != nil {
		ws.Name = *input.Name
	}
	if input.Description != nil {
		ws.Description = input.Description
	}
	if input.IconURL != nil {
		ws.IconURL = input.IconURL
	}
	ws.UpdatedAt = time.Now()

	if err := i.workspaceRepo.Update(ws); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return &UpdateWorkspaceOutput{
		Workspace: WorkspaceOutput{
			ID:          ws.ID,
			Name:        ws.Name,
			Description: ws.Description,
			IconURL:     ws.IconURL,
			Role:        string(member.Role),
			CreatedBy:   ws.CreatedBy,
			CreatedAt:   ws.CreatedAt,
			UpdatedAt:   ws.UpdatedAt,
		},
	}, nil
}

func (i *workspaceInteractor) DeleteWorkspace(input DeleteWorkspaceInput) (*DeleteWorkspaceOutput, error) {
	// Check if user is owner
	member, err := i.workspaceRepo.FindMember(input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil || member.Role != domain.WorkspaceRoleOwner {
		return nil, ErrUnauthorized
	}

	// Delete workspace
	if err := i.workspaceRepo.Delete(input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete workspace: %w", err)
	}

	return &DeleteWorkspaceOutput{Success: true}, nil
}

func (i *workspaceInteractor) ListMembers(input ListMembersInput) (*ListMembersOutput, error) {
	// Check if requester is a member
	member, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// Get all members
	members, err := i.workspaceRepo.FindMembersByWorkspaceID(input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	// Get user details for each member
	output := &ListMembersOutput{
		Members: make([]MemberInfo, 0, len(members)),
	}

	for _, m := range members {
		user, err := i.userRepo.FindByID(m.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user details: %w", err)
		}
		if user == nil {
			continue
		}

		output.Members = append(output.Members, MemberInfo{
			UserID:      user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Role:        string(m.Role),
			JoinedAt:    m.JoinedAt,
		})
	}

	return output, nil
}

func (i *workspaceInteractor) AddMember(input AddMemberInput) (*MemberActionOutput, error) {
	// Check if inviter is admin or owner
	inviterMember, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.InviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check inviter membership: %w", err)
	}
	if inviterMember == nil || (inviterMember.Role != domain.WorkspaceRoleOwner && inviterMember.Role != domain.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	// Validate role
	role := domain.WorkspaceRole(input.Role)
	if !isValidWorkspaceRole(role) {
		return nil, ErrInvalidRole
	}

	// Add member
	member := &domain.WorkspaceMember{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		Role:        role,
		JoinedAt:    time.Now(),
	}

	if err := i.workspaceRepo.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return &MemberActionOutput{Success: true}, nil
}

func (i *workspaceInteractor) UpdateMemberRole(input UpdateMemberRoleInput) (*MemberActionOutput, error) {
	// Check if updater is owner
	updaterMember, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.UpdaterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check updater membership: %w", err)
	}
	if updaterMember == nil || updaterMember.Role != domain.WorkspaceRoleOwner {
		return nil, ErrUnauthorized
	}

	// Check if target member exists
	targetMember, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check target membership: %w", err)
	}
	if targetMember == nil {
		return nil, ErrWorkspaceNotFound
	}

	// Cannot change owner role
	if targetMember.Role == domain.WorkspaceRoleOwner {
		return nil, ErrCannotChangeOwnerRole
	}

	// Validate new role
	role := domain.WorkspaceRole(input.Role)
	if !isValidWorkspaceRole(role) || role == domain.WorkspaceRoleOwner {
		return nil, ErrInvalidRole
	}

	// Update role
	if err := i.workspaceRepo.UpdateMemberRole(input.WorkspaceID, input.UserID, role); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	return &MemberActionOutput{Success: true}, nil
}

func (i *workspaceInteractor) RemoveMember(input RemoveMemberInput) (*MemberActionOutput, error) {
	// Check if remover is admin or owner
	removerMember, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.RemoverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check remover membership: %w", err)
	}
	if removerMember == nil || (removerMember.Role != domain.WorkspaceRoleOwner && removerMember.Role != domain.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	// Check if target member exists
	targetMember, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check target membership: %w", err)
	}
	if targetMember == nil {
		return nil, ErrWorkspaceNotFound
	}

	// Cannot remove owner
	if targetMember.Role == domain.WorkspaceRoleOwner {
		return nil, ErrCannotRemoveOwner
	}

	// Admin can only remove members and guests, not other admins
	if removerMember.Role == domain.WorkspaceRoleAdmin && targetMember.Role == domain.WorkspaceRoleAdmin {
		return nil, ErrUnauthorized
	}

	// Remove member
	if err := i.workspaceRepo.RemoveMember(input.WorkspaceID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	return &MemberActionOutput{Success: true}, nil
}

// Helper function to validate workspace role
func isValidWorkspaceRole(role domain.WorkspaceRole) bool {
	switch role {
	case domain.WorkspaceRoleOwner, domain.WorkspaceRoleAdmin, domain.WorkspaceRoleMember, domain.WorkspaceRoleGuest:
		return true
	default:
		return false
	}
}
