package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

var (
	ErrWorkspaceNotFound     = errors.New("ワークスペースが見つかりません")
	ErrUnauthorized          = errors.New("この操作を行う権限がありません")
	ErrInvalidRole           = errors.New("無効なワークスペースロールです")
	ErrCannotRemoveOwner     = errors.New("ワークスペースのオーナーは削除できません")
	ErrCannotChangeOwnerRole = errors.New("オーナーのロールは変更できません")
)

type WorkspaceUseCase interface {
	GetWorkspacesByUserID(ctx context.Context, userID string) (*GetWorkspacesOutput, error)
	GetWorkspace(ctx context.Context, input GetWorkspaceInput) (*GetWorkspaceOutput, error)
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error)
	UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (*UpdateWorkspaceOutput, error)
	DeleteWorkspace(ctx context.Context, input DeleteWorkspaceInput) (*DeleteWorkspaceOutput, error)
	ListMembers(ctx context.Context, input ListMembersInput) (*ListMembersOutput, error)
	AddMember(ctx context.Context, input AddMemberInput) (*MemberActionOutput, error)
	UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (*MemberActionOutput, error)
	RemoveMember(ctx context.Context, input RemoveMemberInput) (*MemberActionOutput, error)
}

type workspaceInteractor struct {
	workspaceRepo domainrepository.WorkspaceRepository
	userRepo      domainrepository.UserRepository
}

func NewWorkspaceInteractor(
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
) WorkspaceUseCase {
	return &workspaceInteractor{
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (i *workspaceInteractor) GetWorkspacesByUserID(ctx context.Context, userID string) (*GetWorkspacesOutput, error) {
	workspaces, err := i.workspaceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}

	output := &GetWorkspacesOutput{
		Workspaces: make([]WorkspaceOutput, 0, len(workspaces)),
	}

	for _, ws := range workspaces {
		member, err := i.workspaceRepo.FindMember(ctx, ws.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get member info: %w", err)
		}
		if member == nil {
			continue
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

func (i *workspaceInteractor) GetWorkspace(ctx context.Context, input GetWorkspaceInput) (*GetWorkspaceOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	ws, err := i.workspaceRepo.FindByID(ctx, input.ID)
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

func (i *workspaceInteractor) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*CreateWorkspaceOutput, error) {
	workspace := &entity.Workspace{
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.workspaceRepo.Create(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	member := &entity.WorkspaceMember{
		WorkspaceID: workspace.ID,
		UserID:      input.CreatedBy,
		Role:        entity.WorkspaceRoleOwner,
		JoinedAt:    time.Now(),
	}

	if err := i.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add creator as owner: %w", err)
	}

	return &CreateWorkspaceOutput{
		Workspace: WorkspaceOutput{
			ID:          workspace.ID,
			Name:        workspace.Name,
			Description: workspace.Description,
			IconURL:     workspace.IconURL,
			Role:        string(entity.WorkspaceRoleOwner),
			CreatedBy:   workspace.CreatedBy,
			CreatedAt:   workspace.CreatedAt,
			UpdatedAt:   workspace.UpdatedAt,
		},
	}, nil
}

func (i *workspaceInteractor) UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (*UpdateWorkspaceOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil || (member.Role != entity.WorkspaceRoleOwner && member.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	ws, err := i.workspaceRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

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

	if err := i.workspaceRepo.Update(ctx, ws); err != nil {
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

func (i *workspaceInteractor) DeleteWorkspace(ctx context.Context, input DeleteWorkspaceInput) (*DeleteWorkspaceOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.ID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil || member.Role != entity.WorkspaceRoleOwner {
		return nil, ErrUnauthorized
	}

	if err := i.workspaceRepo.Delete(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete workspace: %w", err)
	}

	return &DeleteWorkspaceOutput{Success: true}, nil
}

func (i *workspaceInteractor) ListMembers(ctx context.Context, input ListMembersInput) (*ListMembersOutput, error) {
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RequesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	members, err := i.workspaceRepo.FindMembersByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}

	// ユーザーIDを収集
	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	// ユーザー情報を一括取得
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*entity.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	output := &ListMembersOutput{Members: make([]MemberInfo, 0, len(members))}
	for _, m := range members {
		user := userMap[m.UserID]
		memberInfo := MemberInfo{
			UserID:   m.UserID,
			Role:     string(m.Role),
			JoinedAt: m.JoinedAt,
		}
		if user != nil {
			memberInfo.Email = user.Email
			memberInfo.DisplayName = user.DisplayName
			memberInfo.AvatarURL = user.AvatarURL
		}
		output.Members = append(output.Members, memberInfo)
	}

	return output, nil
}

func (i *workspaceInteractor) AddMember(ctx context.Context, input AddMemberInput) (*MemberActionOutput, error) {
	requester, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.InviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester membership: %w", err)
	}
	if requester == nil || (requester.Role != entity.WorkspaceRoleOwner && requester.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	user, err := i.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", input.UserID)
	}

	if err := validateWorkspaceRole(input.Role); err != nil {
		return nil, err
	}

	member := &entity.WorkspaceMember{
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		Role:        entity.WorkspaceRole(input.Role),
		JoinedAt:    time.Now(),
	}

	if err := i.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return &MemberActionOutput{Success: true}, nil
}

func (i *workspaceInteractor) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (*MemberActionOutput, error) {
	requester, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UpdaterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester membership: %w", err)
	}
	if requester == nil || (requester.Role != entity.WorkspaceRoleOwner && requester.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	if input.UserID == input.UpdaterID && requester.Role == entity.WorkspaceRoleOwner {
		return nil, ErrCannotChangeOwnerRole
	}

	if err := validateWorkspaceRole(input.Role); err != nil {
		return nil, err
	}

	if err := i.workspaceRepo.UpdateMemberRole(ctx, input.WorkspaceID, input.UserID, entity.WorkspaceRole(input.Role)); err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	return &MemberActionOutput{Success: true}, nil
}

func (i *workspaceInteractor) RemoveMember(ctx context.Context, input RemoveMemberInput) (*MemberActionOutput, error) {
	requester, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.RemoverID)
	if err != nil {
		return nil, fmt.Errorf("failed to check requester membership: %w", err)
	}
	if requester == nil || (requester.Role != entity.WorkspaceRoleOwner && requester.Role != entity.WorkspaceRoleAdmin) {
		return nil, ErrUnauthorized
	}

	target, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target member: %w", err)
	}
	if target == nil {
		return nil, ErrWorkspaceNotFound
	}
	if target.Role == entity.WorkspaceRoleOwner {
		return nil, ErrCannotRemoveOwner
	}

	if err := i.workspaceRepo.RemoveMember(ctx, input.WorkspaceID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	return &MemberActionOutput{Success: true}, nil
}

func validateWorkspaceRole(role string) error {
	switch entity.WorkspaceRole(role) {
	case entity.WorkspaceRoleOwner, entity.WorkspaceRoleAdmin, entity.WorkspaceRoleMember, entity.WorkspaceRoleGuest:
		return nil
	default:
		return ErrInvalidRole
	}
}
