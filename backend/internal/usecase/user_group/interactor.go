package user_group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

var (
	ErrUserGroupNotFound   = errors.New("ユーザーグループが見つかりません")
	ErrUnauthorized        = errors.New("この操作を行う権限がありません")
	ErrUserGroupNameExists = errors.New("同じ名前のユーザーグループが既に存在します")
	ErrUserAlreadyInGroup  = errors.New("ユーザーは既にこのグループに参加しています")
	ErrUserNotInGroup      = errors.New("ユーザーはこのグループに参加していません")
)

type UserGroupUseCase interface {
	CreateUserGroup(ctx context.Context, input CreateUserGroupInput) (*CreateUserGroupOutput, error)
	UpdateUserGroup(ctx context.Context, input UpdateUserGroupInput) (*UpdateUserGroupOutput, error)
	DeleteUserGroup(ctx context.Context, input DeleteUserGroupInput) (*DeleteUserGroupOutput, error)
	GetUserGroup(ctx context.Context, input GetUserGroupInput) (*GetUserGroupOutput, error)
	ListUserGroups(ctx context.Context, input ListUserGroupsInput) (*ListUserGroupsOutput, error)
	AddMember(ctx context.Context, input AddMemberInput) (*AddMemberOutput, error)
	RemoveMember(ctx context.Context, input RemoveMemberInput) (*RemoveMemberOutput, error)
	ListMembers(ctx context.Context, input ListMembersInput) (*ListMembersOutput, error)
}

type userGroupInteractor struct {
	userGroupRepo domainrepository.UserGroupRepository
	workspaceRepo domainrepository.WorkspaceRepository
	userRepo      domainrepository.UserRepository
}

func NewUserGroupInteractor(
	userGroupRepo domainrepository.UserGroupRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
) UserGroupUseCase {
	return &userGroupInteractor{
		userGroupRepo: userGroupRepo,
		workspaceRepo: workspaceRepo,
		userRepo:      userRepo,
	}
}

func (i *userGroupInteractor) CreateUserGroup(ctx context.Context, input CreateUserGroupInput) (*CreateUserGroupOutput, error) {
	// ワークスペースの存在確認と権限チェック
	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, errors.New("ワークスペースが見つかりません")
	}

	// 作成者がワークスペースのメンバーかチェック
	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// グループ名の重複チェック
	existing, err := i.userGroupRepo.FindByName(ctx, input.WorkspaceID, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check group name: %w", err)
	}
	if existing != nil {
		return nil, ErrUserGroupNameExists
	}

	// グループ作成
	group := &entity.UserGroup{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		CreatedBy:   input.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := i.userGroupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create user group: %w", err)
	}

	output := toUserGroupOutput(group)
	return &CreateUserGroupOutput{UserGroup: output}, nil
}

func (i *userGroupInteractor) UpdateUserGroup(ctx context.Context, input UpdateUserGroupInput) (*UpdateUserGroupOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(ctx, input.ID)
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
		existing, err := i.userGroupRepo.FindByName(ctx, group.WorkspaceID, *input.Name)
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

	if err := i.userGroupRepo.Update(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to update user group: %w", err)
	}

	output := toUserGroupOutput(group)
	return &UpdateUserGroupOutput{UserGroup: output}, nil
}

func (i *userGroupInteractor) DeleteUserGroup(ctx context.Context, input DeleteUserGroupInput) (*DeleteUserGroupOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(ctx, input.ID)
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

	if err := i.userGroupRepo.Delete(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete user group: %w", err)
	}

	return &DeleteUserGroupOutput{Success: true}, nil
}

func (i *userGroupInteractor) GetUserGroup(ctx context.Context, input GetUserGroupInput) (*GetUserGroupOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（ワークスペースメンバーのみアクセス可能）
	member, err := i.workspaceRepo.FindMember(ctx, group.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	output := toUserGroupOutput(group)
	return &GetUserGroupOutput{UserGroup: output}, nil
}

func (i *userGroupInteractor) ListUserGroups(ctx context.Context, input ListUserGroupsInput) (*ListUserGroupsOutput, error) {
	// ワークスペースの存在確認と権限チェック
	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, errors.New("ワークスペースが見つかりません")
	}

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// グループ一覧取得
	groups, err := i.userGroupRepo.FindByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user groups: %w", err)
	}

	outputs := make([]UserGroupOutput, len(groups))
	for i, group := range groups {
		outputs[i] = toUserGroupOutput(group)
	}

	return &ListUserGroupsOutput{UserGroups: outputs}, nil
}

func (i *userGroupInteractor) AddMember(ctx context.Context, input AddMemberInput) (*AddMemberOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(ctx, input.GroupID)
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
	isMember, err := i.userGroupRepo.IsMember(ctx, input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return nil, ErrUserAlreadyInGroup
	}

	// メンバー追加
	member := &entity.UserGroupMember{
		GroupID:  input.GroupID,
		UserID:   input.UserID,
		JoinedAt: time.Now(),
	}

	if err := i.userGroupRepo.AddMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	return &AddMemberOutput{Success: true}, nil
}

func (i *userGroupInteractor) RemoveMember(ctx context.Context, input RemoveMemberInput) (*RemoveMemberOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(ctx, input.GroupID)
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
	isMember, err := i.userGroupRepo.IsMember(ctx, input.GroupID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return nil, ErrUserNotInGroup
	}

	// メンバー削除
	if err := i.userGroupRepo.RemoveMember(ctx, input.GroupID, input.UserID); err != nil {
		return nil, fmt.Errorf("failed to remove member: %w", err)
	}

	return &RemoveMemberOutput{Success: true}, nil
}

func (i *userGroupInteractor) ListMembers(ctx context.Context, input ListMembersInput) (*ListMembersOutput, error) {
	// グループの存在確認
	group, err := i.userGroupRepo.FindByID(ctx, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to load user group: %w", err)
	}
	if group == nil {
		return nil, ErrUserGroupNotFound
	}

	// 権限チェック（ワークスペースメンバーのみアクセス可能）
	member, err := i.workspaceRepo.FindMember(ctx, group.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, ErrUnauthorized
	}

	// メンバー一覧取得
	groupMembers, err := i.userGroupRepo.FindMembersByGroupID(ctx, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group members: %w", err)
	}

	// ユーザー情報を取得
	userIDs := make([]string, len(groupMembers))
	for i, member := range groupMembers {
		userIDs[i] = member.UserID
	}

	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	// ユーザー情報をマップに格納
	userMap := make(map[string]*entity.User)
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

func toUserGroupOutput(group *entity.UserGroup) UserGroupOutput {
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
