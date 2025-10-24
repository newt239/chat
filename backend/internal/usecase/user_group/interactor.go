package user_group

import (
	"errors"
	"fmt"
	"time"

	"github.com/example/chat/internal/domain"
)

var (
	ErrUserGroupNotFound     = errors.New("user group not found")
	ErrUnauthorized          = errors.New("unauthorized to perform this action")
	ErrUserGroupNameExists   = errors.New("user group name already exists")
	ErrUserAlreadyInGroup    = errors.New("user is already a member of this group")
	ErrUserNotInGroup        = errors.New("user is not a member of this group")
)

type UserGroupUseCase interface {
	CreateUserGroup(input CreateUserGroupInput) (*CreateUserGroupOutput, error)
	UpdateUserGroup(input UpdateUserGroupInput) (*UpdateUserGroupOutput, error)
	DeleteUserGroup(input DeleteUserGroupInput) (*DeleteUserGroupOutput, error)
	GetUserGroup(input GetUserGroupInput) (*GetUserGroupOutput, error)
	ListUserGroups(input ListUserGroupsInput) (*ListUserGroupsOutput, error)
	AddMember(input AddMemberInput) (*AddMemberOutput, error)
	RemoveMember(input RemoveMemberInput) (*RemoveMemberOutput, error)
	ListMembers(input ListMembersInput) (*ListMembersOutput, error)
}

type userGroupInteractor struct {
	userGroupRepo domain.UserGroupRepository
	workspaceRepo domain.WorkspaceRepository
	userRepo      domain.UserRepository
}

func NewUserGroupInteractor(
	userGroupRepo domain.UserGroupRepository,
	workspaceRepo domain.WorkspaceRepository,
	userRepo domain.UserRepository,
) UserGroupUseCase {
	return &userGroupInteractor{
		userGroupRepo: userGroupRepo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (i *userGroupInteractor) CreateUserGroup(input CreateUserGroupInput) (*CreateUserGroupOutput, error) {
	// ワークスペースの存在確認と権限チェック
	workspace, err := i.workspaceRepo.FindByID(input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, errors.New("workspace not found")
	}

	// 作成者がワークスペースのメンバーかチェック
	member, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// グループ名の重複チェック
	existing, err := i.userGroupRepo.FindByName(input.WorkspaceID, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check group name: %w", err)
	}
	if existing != nil {
		return nil, ErrUserGroupNameExists
	}

	// グループ作成
	group := &domain.UserGroup{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.userGroupRepo.Create(group); err != nil {
		return nil, fmt.Errorf("failed to create user group: %w", err)
	}

	output := toUserGroupOutput(group)
	return &CreateUserGroupOutput{UserGroup: output}, nil
}

func (i *userGroupInteractor) UpdateUserGroup(input UpdateUserGroupInput) (*UpdateUserGroupOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（作成者のみ更新可能）
	if group.CreatedBy != input.UpdatedBy {
		return nil, ErrUnauthorized
	}

	// 名前の更新がある場合は重複チェック
	if input.Name != nil && *input.Name != group.Name {
		existing, err := i.userGroupRepo.FindByName(group.WorkspaceID, *input.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check group name: %w", err)
		}
		if existing != nil {
			return nil, ErrUserGroupNameExists
		}
		group.Name = *input.Name
	}

	if input.Description != nil {
		group.Description = input.Description
	}

	group.UpdatedAt = time.Now()

	if err := i.userGroupRepo.Update(group); err != nil {
		return nil, fmt.Errorf("failed to update user group: %w", err)
	}

	output := toUserGroupOutput(group)
	return &UpdateUserGroupOutput{UserGroup: output}, nil
}

func (i *userGroupInteractor) DeleteUserGroup(input DeleteUserGroupInput) (*DeleteUserGroupOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（作成者のみ削除可能）
	if group.CreatedBy != input.DeletedBy {
		return nil, ErrUnauthorized
	}

	if err := i.userGroupRepo.Delete(input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete user group: %w", err)
	}

	return &DeleteUserGroupOutput{Success: true}, nil
}

func (i *userGroupInteractor) GetUserGroup(input GetUserGroupInput) (*GetUserGroupOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（ワークスペースメンバーのみアクセス可能）
	member, err := i.workspaceRepo.FindMember(group.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	output := toUserGroupOutput(group)
	return &GetUserGroupOutput{UserGroup: output}, nil
}

func (i *userGroupInteractor) ListUserGroups(input ListUserGroupsInput) (*ListUserGroupsOutput, error) {
	// ワークスペースの存在確認と権限チェック
	workspace, err := i.workspaceRepo.FindByID(input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, errors.New("workspace not found")
	}

	member, err := i.workspaceRepo.FindMember(input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// グループ一覧取得
	groups, err := i.userGroupRepo.FindByWorkspaceID(input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user groups: %w", err)
	}

	outputs := make([]UserGroupOutput, len(groups))
	for i, group := range groups {
		outputs[i] = toUserGroupOutput(group)
	}

	return &ListUserGroupsOutput{UserGroups: outputs}, nil
}

func (i *userGroupInteractor) AddMember(input AddMemberInput) (*AddMemberOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（作成者のみメンバー追加可能）
	if group.CreatedBy != input.AddedBy {
		return nil, ErrUnauthorized
	}

	// 既にメンバーかチェック
	isMember, err := i.userGroupRepo.IsMember(input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return nil, ErrUserAlreadyInGroup
	}

	// メンバー追加
	member := &domain.UserGroupMember{
		GroupID:  input.GroupID,
		UserID:   input.UserID,
		JoinedAt: time.Now(),
	}

	if err := i.userGroupRepo.AddMember(member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return &AddMemberOutput{Success: true}, nil
}

func (i *userGroupInteractor) RemoveMember(input RemoveMemberInput) (*RemoveMemberOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（作成者または本人のみ削除可能）
	if group.CreatedBy != input.RemovedBy && input.UserID != input.RemovedBy {
		return nil, ErrUnauthorized
	}

	// メンバーかチェック
	isMember, err := i.userGroupRepo.IsMember(input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrUserNotInGroup
	}

	// メンバー削除
	if err := i.userGroupRepo.RemoveMember(input.GroupID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	return &RemoveMemberOutput{Success: true}, nil
}

func (i *userGroupInteractor) ListMembers(input ListMembersInput) (*ListMembersOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（ワークスペースメンバーのみアクセス可能）
	member, err := i.workspaceRepo.FindMember(group.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// メンバー一覧取得
	groupMembers, err := i.userGroupRepo.FindMembersByGroupID(input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group members: %w", err)
	}

	// ユーザー情報を取得
	userIDs := make([]string, len(groupMembers))
	for i, member := range groupMembers {
		userIDs[i] = member.UserID
	}

	users, err := i.userRepo.FindByIDs(userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*domain.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	// メンバー情報を構築
	members := make([]MemberInfo, 0, len(groupMembers))
	for _, member := range groupMembers {
		user := userMap[member.UserID]
		if user != nil {
			members = append(members, MemberInfo{
				UserID:      user.ID,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
				JoinedAt:    member.JoinedAt,
			})
		}
	}

	return &ListMembersOutput{Members: members}, nil
}

func toUserGroupOutput(group *domain.UserGroup) UserGroupOutput {
	return UserGroupOutput{
		ID:          group.ID,
		WorkspaceID: group.WorkspaceID,
		Name:        group.Name,
		Description: group.Description,
		CreatedBy:   group.CreatedBy,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}
