package channelmember

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

var (
	ErrUnauthorized     = errors.New("この操作を行う権限がありません")
	ErrChannelNotFound  = errors.New("チャンネルが見つかりません")
	ErrUserNotFound     = errors.New("ユーザーが見つかりません")
	ErrAlreadyMember    = errors.New("ユーザーは既にメンバーです")
	ErrNotMember        = errors.New("ユーザーはメンバーではありません")
	ErrInvalidRole      = errors.New("無効なロールです")
	ErrChannelNotPublic = errors.New("このチャンネルは公開されていません")
	ErrLastAdminRemoval = errors.New("最後の管理者は削除できません")
)

type ChannelMemberUseCase interface {
	ListMembers(ctx context.Context, input ListMembersInput) (*MemberListOutput, error)
	InviteMember(ctx context.Context, input InviteMemberInput) error
	JoinPublicChannel(ctx context.Context, input JoinChannelInput) error
	UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) error
	RemoveMember(ctx context.Context, input RemoveMemberInput) error
	LeaveChannel(ctx context.Context, input LeaveChannelInput) error
}

type channelMemberInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	userRepo          domainrepository.UserRepository
}

func NewChannelMemberInteractor(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
) ChannelMemberUseCase {
	return &channelMemberInteractor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
	}
}

func (i *channelMemberInteractor) ListMembers(ctx context.Context, input ListMembersInput) (*MemberListOutput, error) {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return nil, err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return nil, err
	}

	channel, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	// プライベートチャンネルの場合、アクセス権を確認
	if channel.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to check membership: %w", err)
		}
		if !isMember {
			return nil, ErrUnauthorized
		}
	} else {
		// パブリックチャンネルの場合、ワークスペースメンバーかどうか確認
		member, err := i.workspaceRepo.FindMember(ctx, channel.WorkspaceID, input.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify workspace membership: %w", err)
		}
		if member == nil {
			return nil, ErrUnauthorized
		}
	}

	members, err := i.channelMemberRepo.FindMembers(ctx, input.ChannelID)
	if err != nil {
		return nil, fmt.Errorf("failed to find members: %w", err)
	}

	userIDs := make([]string, len(members))
	for idx, m := range members {
		userIDs[idx] = m.UserID
	}

	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}

	userMap := make(map[string]*entity.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	memberInfos := make([]MemberInfo, 0, len(members))
	for _, m := range members {
		user := userMap[m.UserID]
		if user == nil {
			continue
		}
		memberInfos = append(memberInfos, MemberInfo{
			UserID:      m.UserID,
			Role:        string(m.Role),
			JoinedAt:    m.JoinedAt,
			DisplayName: user.DisplayName,
			Email:       user.Email,
			AvatarURL:   user.AvatarURL,
		})
	}

	return &MemberListOutput{Members: memberInfos}, nil
}

func (i *channelMemberInteractor) InviteMember(ctx context.Context, input InviteMemberInput) error {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return err
	}
	if err := validateUUID(input.OperatorID, "operator user ID"); err != nil {
		return err
	}
	if err := validateUUID(input.TargetUserID, "target user ID"); err != nil {
		return err
	}

	// ロールの検証
	role := entity.ChannelRole(input.Role)
	if role != entity.ChannelRoleMember && role != entity.ChannelRoleAdmin {
		return ErrInvalidRole
	}

	channel, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// オペレーターがワークスペースメンバーかどうか確認
	operatorMember, err := i.workspaceRepo.FindMember(ctx, channel.WorkspaceID, input.OperatorID)
	if err != nil {
		return fmt.Errorf("failed to verify operator workspace membership: %w", err)
	}
	if operatorMember == nil {
		return ErrUnauthorized
	}

	// プライベートチャンネルの場合、オペレーターのアクセス権を確認
	if channel.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.OperatorID)
		if err != nil {
			return fmt.Errorf("failed to check operator membership: %w", err)
		}
		if !isMember {
			return ErrUnauthorized
		}
	}

	// 招待・削除・ロール変更を実行できるのはワークスペースの owner/admin とチャンネル作成者
	if operatorMember.Role != entity.WorkspaceRoleOwner &&
		operatorMember.Role != entity.WorkspaceRoleAdmin &&
		channel.CreatedBy != input.OperatorID {
		return ErrUnauthorized
	}

	// ターゲットユーザーがワークスペースメンバーであるか検証
	targetMember, err := i.workspaceRepo.FindMember(ctx, channel.WorkspaceID, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("failed to verify target user workspace membership: %w", err)
	}
	if targetMember == nil {
		return ErrUserNotFound
	}

	// 既存メンバーなら409エラー
	isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("failed to check target user membership: %w", err)
	}
	if isMember {
		return ErrAlreadyMember
	}

	member := &entity.ChannelMember{
		ChannelID: input.ChannelID,
		UserID:    input.TargetUserID,
		Role:      role,
		JoinedAt:  time.Now(),
	}

	if err := i.channelMemberRepo.AddMember(ctx, member); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

func (i *channelMemberInteractor) JoinPublicChannel(ctx context.Context, input JoinChannelInput) error {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return err
	}

	channel, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// 対象チャンネルがパブリックであることを確認
	if channel.IsPrivate {
		return ErrChannelNotPublic
	}

	// ワークスペースメンバーであることを確認
	member, err := i.workspaceRepo.FindMember(ctx, channel.WorkspaceID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to verify workspace membership: %w", err)
	}
	if member == nil {
		return ErrUnauthorized
	}

	// 既存メンバーの場合は冪等に成功応答
	isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return nil
	}

	channelMember := &entity.ChannelMember{
		ChannelID: input.ChannelID,
		UserID:    input.UserID,
		Role:      entity.ChannelRoleMember,
		JoinedAt:  time.Now(),
	}

	if err := i.channelMemberRepo.AddMember(ctx, channelMember); err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}

	return nil
}

func (i *channelMemberInteractor) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) error {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return err
	}
	if err := validateUUID(input.OperatorID, "operator user ID"); err != nil {
		return err
	}
	if err := validateUUID(input.TargetUserID, "target user ID"); err != nil {
		return err
	}

	// ロールの検証
	role := entity.ChannelRole(input.Role)
	if role != entity.ChannelRoleMember && role != entity.ChannelRoleAdmin {
		return ErrInvalidRole
	}

	channel, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// オペレーターがワークスペースメンバーかどうか確認
	operatorMember, err := i.workspaceRepo.FindMember(ctx, channel.WorkspaceID, input.OperatorID)
	if err != nil {
		return fmt.Errorf("failed to verify operator workspace membership: %w", err)
	}
	if operatorMember == nil {
		return ErrUnauthorized
	}

	// プライベートチャンネルの場合、オペレーターのアクセス権を確認
	if channel.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.OperatorID)
		if err != nil {
			return fmt.Errorf("failed to check operator membership: %w", err)
		}
		if !isMember {
			return ErrUnauthorized
		}
	}

	// 招待・削除・ロール変更を実行できるのはワークスペースの owner/admin とチャンネル作成者
	if operatorMember.Role != entity.WorkspaceRoleOwner &&
		operatorMember.Role != entity.WorkspaceRoleAdmin &&
		channel.CreatedBy != input.OperatorID {
		return ErrUnauthorized
	}

	// 対象ユーザーがチャンネルメンバーであることを確認
	isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("failed to check target user membership: %w", err)
	}
	if !isMember {
		return ErrNotMember
	}

	// ロールをmemberに降格する場合、管理者が最低1名残るか検証
	members, err := i.channelMemberRepo.FindMembers(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find members: %w", err)
	}

	// 現在のターゲットユーザーのロールを確認
	var currentRole entity.ChannelRole
	for _, m := range members {
		if m.UserID == input.TargetUserID {
			currentRole = m.Role
			break
		}
	}

	// 管理者から一般メンバーに降格する場合
	if currentRole == entity.ChannelRoleAdmin && role == entity.ChannelRoleMember {
		adminCount, err := i.channelMemberRepo.CountAdmins(ctx, input.ChannelID)
		if err != nil {
			return fmt.Errorf("failed to count admins: %w", err)
		}
		if adminCount <= 1 {
			return ErrLastAdminRemoval
		}
	}

	if err := i.channelMemberRepo.UpdateMemberRole(ctx, input.ChannelID, input.TargetUserID, role); err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	return nil
}

func (i *channelMemberInteractor) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return err
	}
	if err := validateUUID(input.OperatorID, "operator user ID"); err != nil {
		return err
	}
	if err := validateUUID(input.TargetUserID, "target user ID"); err != nil {
		return err
	}

	channel, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// オペレーターがワークスペースメンバーかどうか確認
	operatorMember, err := i.workspaceRepo.FindMember(ctx, channel.WorkspaceID, input.OperatorID)
	if err != nil {
		return fmt.Errorf("failed to verify operator workspace membership: %w", err)
	}
	if operatorMember == nil {
		return ErrUnauthorized
	}

	// プライベートチャンネルの場合、オペレーターのアクセス権を確認
	if channel.IsPrivate {
		isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.OperatorID)
		if err != nil {
			return fmt.Errorf("failed to check operator membership: %w", err)
		}
		if !isMember {
			return ErrUnauthorized
		}
	}

	// 招待・削除・ロール変更を実行できるのはワークスペースの owner/admin とチャンネル作成者
	if operatorMember.Role != entity.WorkspaceRoleOwner &&
		operatorMember.Role != entity.WorkspaceRoleAdmin &&
		channel.CreatedBy != input.OperatorID {
		return ErrUnauthorized
	}

	// 対象ユーザーがメンバーであることを確認
	isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("failed to check target user membership: %w", err)
	}
	if !isMember {
		return ErrNotMember
	}

	// 削除対象が管理者の場合、残りの管理者数を確認
	members, err := i.channelMemberRepo.FindMembers(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find members: %w", err)
	}

	for _, m := range members {
		if m.UserID == input.TargetUserID && m.Role == entity.ChannelRoleAdmin {
			adminCount, err := i.channelMemberRepo.CountAdmins(ctx, input.ChannelID)
			if err != nil {
				return fmt.Errorf("failed to count admins: %w", err)
			}
			if adminCount <= 1 {
				return ErrLastAdminRemoval
			}
			break
		}
	}

	if err := i.channelMemberRepo.RemoveMember(ctx, input.ChannelID, input.TargetUserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

func (i *channelMemberInteractor) LeaveChannel(ctx context.Context, input LeaveChannelInput) error {
	if err := validateUUID(input.ChannelID, "channel ID"); err != nil {
		return err
	}
	if err := validateUUID(input.UserID, "user ID"); err != nil {
		return err
	}

	channel, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if channel == nil {
		return ErrChannelNotFound
	}

	// 当該ユーザーがメンバーであることを確認
	isMember, err := i.channelMemberRepo.IsMember(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return ErrNotMember
	}

	// 離脱者が管理者かつ最後の1人なら離脱不可
	members, err := i.channelMemberRepo.FindMembers(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find members: %w", err)
	}

	for _, m := range members {
		if m.UserID == input.UserID && m.Role == entity.ChannelRoleAdmin {
			adminCount, err := i.channelMemberRepo.CountAdmins(ctx, input.ChannelID)
			if err != nil {
				return fmt.Errorf("failed to count admins: %w", err)
			}
			if adminCount <= 1 {
				return ErrLastAdminRemoval
			}
			break
		}
	}

	if err := i.channelMemberRepo.RemoveMember(ctx, input.ChannelID, input.UserID); err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	return nil
}

func validateUUID(id string, label string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: invalid %s format", domerr.ErrValidation, label)
	}
	return nil
}
