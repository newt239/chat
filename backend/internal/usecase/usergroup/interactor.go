package usergroup

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

var (
	ErrUserGroupNotFound   = domerr.New(domerr.ErrNotFound, "ユーザーグループが見つかりません")
	ErrUserGroupNameExists = domerr.New(domerr.ErrAlreadyExists, "同じ名前のユーザーグループが既に存在します")
	ErrUserNotInGroup      = domerr.New(domerr.ErrFailedPrecondition, "ユーザーはこのグループに参加していません")
)

type CreateInput struct {
	WorkspaceID string
	Name        string
	Description *string
	CreatedBy   string
}

type UpdateInput struct {
	ID          string
	Name        *string
	Description *string
	UpdatedBy   string
}

// MemberInput は OperatorID のユーザーが UserID のユーザーをグループに追加・削除します
type MemberInput struct {
	GroupID    string
	UserID     string
	OperatorID string
}

type Interactor struct {
	userGroupRepo domainrepository.UserGroupRepository
	workspaceRepo domainrepository.WorkspaceRepository
	userRepo      domainrepository.UserRepository
}

func New(
	userGroupRepo domainrepository.UserGroupRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
) *Interactor {
	return &Interactor{userGroupRepo: userGroupRepo, workspaceRepo: workspaceRepo, userRepo: userRepo}
}

func (i *Interactor) findGroup(ctx context.Context, groupID string) (*entity.UserGroup, error) {
	group, err := i.userGroupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}
	return group, nil
}

func (i *Interactor) ensureNameAvailable(ctx context.Context, workspaceID, name string) error {
	existing, err := i.userGroupRepo.FindByName(ctx, workspaceID, name)
	if err != nil {
		return fmt.Errorf("failed to check group name: %w", err)
	}
	if existing != nil {
		return ErrUserGroupNameExists
	}
	return nil
}

// Create は owner / admin だけが実行できます
func (i *Interactor) Create(ctx context.Context, input CreateInput) (*entity.UserGroup, error) {
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.WorkspaceID, input.CreatedBy); err != nil {
		return nil, err
	}
	if err := i.ensureNameAvailable(ctx, input.WorkspaceID, input.Name); err != nil {
		return nil, err
	}
	group := &entity.UserGroup{WorkspaceID: input.WorkspaceID, Name: input.Name, Description: input.Description, CreatedBy: input.CreatedBy}
	if err := i.userGroupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create user group: %w", err)
	}
	return group, nil
}

func (i *Interactor) Update(ctx context.Context, input UpdateInput) (*entity.UserGroup, error) {
	group, err := i.findGroup(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, group.WorkspaceID, input.UpdatedBy); err != nil {
		return nil, err
	}
	if input.Name != nil && *input.Name != group.Name {
		if err := i.ensureNameAvailable(ctx, group.WorkspaceID, *input.Name); err != nil {
			return nil, err
		}
		group.Name = *input.Name
	}
	if input.Description != nil {
		group.Description = input.Description
	}
	if err := i.userGroupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update user group: %w", err)
	}
	return group, nil
}

func (i *Interactor) Delete(ctx context.Context, groupID, userID string) error {
	group, err := i.findGroup(ctx, groupID)
	if err != nil {
		return err
	}
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, group.WorkspaceID, userID); err != nil {
		return err
	}
	if err := i.userGroupRepo.Delete(ctx, groupID); err != nil {
		return fmt.Errorf("failed to delete user group: %w", err)
	}
	return nil
}

func (i *Interactor) List(ctx context.Context, workspaceID, userID string) ([]*entity.UserGroup, error) {
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, workspaceID, userID); err != nil {
		return nil, err
	}
	groups, err := i.userGroupRepo.FindByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user groups: %w", err)
	}
	return groups, nil
}

// AddMember は停止中やワークスペース外のユーザーを追加せず、既に参加していれば ErrAlreadyMember を返します
func (i *Interactor) AddMember(ctx context.Context, input MemberInput) error {
	group, err := i.findGroup(ctx, input.GroupID)
	if err != nil {
		return err
	}
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, group.WorkspaceID, input.OperatorID); err != nil {
		return err
	}
	target, err := i.workspaceRepo.FindMember(ctx, group.WorkspaceID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to verify target membership: %w", err)
	}
	if target == nil {
		return domerr.ErrUserNotFound
	}
	return i.userGroupRepo.AddMember(ctx, &entity.UserGroupMember{GroupID: input.GroupID, UserID: input.UserID})
}

// RemoveMember は本人なら管理者でなくても抜けられます
func (i *Interactor) RemoveMember(ctx context.Context, input MemberInput) error {
	group, err := i.findGroup(ctx, input.GroupID)
	if err != nil {
		return err
	}
	if input.UserID != input.OperatorID {
		if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, group.WorkspaceID, input.OperatorID); err != nil {
			return err
		}
	}
	isMember, err := i.userGroupRepo.IsMember(ctx, input.GroupID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return ErrUserNotInGroup
	}
	if err := i.userGroupRepo.RemoveMember(ctx, input.GroupID, input.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

func (i *Interactor) ListMembers(ctx context.Context, groupID, userID string) ([]*entity.User, error) {
	group, err := i.findGroup(ctx, groupID)
	if err != nil {
		return nil, err
	}
	if _, err := domainservice.EnsureMember(ctx, i.workspaceRepo, group.WorkspaceID, userID); err != nil {
		return nil, err
	}
	groupMembers, err := i.userGroupRepo.FindMembersByGroupIDs(ctx, []string{groupID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group members: %w", err)
	}
	userIDs := make([]string, len(groupMembers))
	for idx, member := range groupMembers {
		userIDs[idx] = member.UserID
	}
	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	members := make([]*entity.User, 0, len(groupMembers))
	for _, member := range groupMembers {
		if user := users[member.UserID]; user != nil {
			members = append(members, user)
		}
	}
	return members, nil
}
