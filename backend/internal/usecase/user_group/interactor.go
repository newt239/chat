package user_group

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

var (
	ErrUserGroupNotFound   = errors.New("ユーザーグループが見つかりません")
	ErrUserGroupNameExists = errors.New("同じ名前のユーザーグループが既に存在します")
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

// ensureCanManage はユーザーグループを編集できる owner / admin であることを確認します。停止中のメンバーは FindMember が返さない
func (i *userGroupInteractor) ensureCanManage(ctx context.Context, workspaceID, userID string) error {
	member, err := i.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if !member.IsAdmin() {
		return domerr.ErrUnauthorized
	}
	return nil
}

func (i *userGroupInteractor) CreateUserGroup(ctx context.Context, input CreateUserGroupInput) (*CreateUserGroupOutput, error) {
	// ワークスペースの存在確認と権限チェック
	workspace, err := i.workspaceRepo.FindByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workspace: %w", err)
	}
	if workspace == nil {
		return nil, domerr.ErrWorkspaceNotFound
	}

	if err := i.ensureCanManage(ctx, input.WorkspaceID, input.CreatedBy); err != nil {
		return nil, err
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

	if err := i.ensureCanManage(ctx, group.WorkspaceID, input.UpdatedBy); err != nil {
		return nil, err
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

	if err := i.ensureCanManage(ctx, group.WorkspaceID, input.DeletedBy); err != nil {
		return nil, err
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
		return nil, domerr.ErrUnauthorized
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
		return nil, domerr.ErrWorkspaceNotFound
	}

	member, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return nil, domerr.ErrUnauthorized
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

	if err := i.ensureCanManage(ctx, group.WorkspaceID, input.AddedBy); err != nil {
		return nil, err
	}

	// 停止中やワークスペース外のユーザーはグループに入れない
	target, err := i.workspaceRepo.FindMember(ctx, group.WorkspaceID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify target membership: %w", err)
	}
	if target == nil {
		return nil, domerr.ErrUserNotFound
	}

	// 既に参加していれば ErrAlreadyMember になる
	if err := i.userGroupRepo.AddMember(ctx, &entity.UserGroupMember{GroupID: input.GroupID, UserID: input.UserID, JoinedAt: time.Now()}); err != nil {
		return nil, err
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

	// 本人はグループから抜けられる
	if input.UserID != input.RemovedBy {
		if err := i.ensureCanManage(ctx, group.WorkspaceID, input.RemovedBy); err != nil {
			return nil, err
		}
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
		return nil, domerr.ErrUnauthorized
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
