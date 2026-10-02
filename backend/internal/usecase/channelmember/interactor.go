package channelmember

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

var (
	ErrNotMember        = errors.New("ユーザーはメンバーではありません")
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

// ChannelRevoker はチャンネルから外れたユーザーへのリアルタイム配信を止めます
type ChannelRevoker interface {
	RevokeChannel(workspaceID, channelID, userID string)
}

type channelMemberInteractor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	userRepo          domainrepository.UserRepository
	systemMessageUC   systemmessage.UseCase
	channelAccessSvc  service.ChannelAccessService
	txManager         domaintransaction.Manager
	revoker           ChannelRevoker
	logger            service.Logger
}

func NewChannelMemberInteractor(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	systemMessageUC systemmessage.UseCase,
	channelAccessSvc service.ChannelAccessService,
	txManager domaintransaction.Manager,
	revoker ChannelRevoker,
	logger service.Logger,
) ChannelMemberUseCase {
	return &channelMemberInteractor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
		systemMessageUC:   systemMessageUC,
		channelAccessSvc:  channelAccessSvc,
		txManager:         txManager,
		revoker:           revoker,
		logger:            logger,
	}
}

// recordSystemMessage はメンバーの増減をチャンネルのタイムラインに残します。失敗してもメンバーの変更は取り消さない
func (i *channelMemberInteractor) recordSystemMessage(ctx context.Context, ch *entity.Channel, kind entity.SystemMessageKind, actorID, targetUserID string) {
	if _, err := i.systemMessageUC.Create(ctx, systemmessage.CreateInput{
		Channel:   ch,
		Kind:      kind,
		Payload:   map[string]any{"actorId": actorID, "userId": targetUserID},
		ActorID:   &actorID,
	}); err != nil {
		i.logger.Warn("メンバーの変更をタイムラインに残せません", service.LogField{Key: "channelId", Value: ch.ID}, service.LogField{Key: "error", Value: err.Error()})
	}
}

// ensureCanManageMembers は招待・削除・ロール変更を行えるのがワークスペースの管理者とチャンネルの作成者だけであることを確かめます
func (i *channelMemberInteractor) ensureCanManageMembers(ctx context.Context, channelID, operatorID string) (*entity.Channel, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, operatorID)
	if err != nil {
		return nil, err
	}
	if ch.CreatedBy == operatorID {
		return ch, nil
	}
	operator, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify operator workspace membership: %w", err)
	}
	if !operator.IsAdmin() {
		return nil, domerr.ErrUnauthorized
	}
	return ch, nil
}

// changeMember は対象者のロールをトランザクションの中で 1 件だけ読み、最後の管理者がいなくなる変更を拒否してから change を実行します
// newRole が nil のときは対象者を外す変更として扱います
func (i *channelMemberInteractor) changeMember(ctx context.Context, channelID, userID string, newRole *entity.ChannelRole, change func(ctx context.Context) error) error {
	return i.txManager.Do(ctx, func(ctx context.Context) error {
		member, err := i.channelMemberRepo.FindMember(ctx, channelID, userID)
		if err != nil {
			return fmt.Errorf("failed to find member: %w", err)
		}
		if member == nil {
			return ErrNotMember
		}
		if member.Role == entity.ChannelRoleAdmin && (newRole == nil || *newRole != entity.ChannelRoleAdmin) {
			admins, err := i.channelMemberRepo.CountAdmins(ctx, channelID)
			if err != nil {
				return fmt.Errorf("failed to count admins: %w", err)
			}
			if admins <= 1 {
				return ErrLastAdminRemoval
			}
		}
		return change(ctx)
	})
}

func parseRole(role string) (entity.ChannelRole, error) {
	switch r := entity.ChannelRole(role); r {
	case entity.ChannelRoleMember, entity.ChannelRoleAdmin:
		return r, nil
	}
	return "", domerr.ErrInvalidRole
}

func (i *channelMemberInteractor) ListMembers(ctx context.Context, input ListMembersInput) (*MemberListOutput, error) {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, input.ChannelID, input.UserID); err != nil {
		return nil, err
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
	userMap := make(map[string]*entity.User, len(users))
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

// InviteMember は既に参加していれば ErrAlreadyMember を返します
func (i *channelMemberInteractor) InviteMember(ctx context.Context, input InviteMemberInput) error {
	role, err := parseRole(input.Role)
	if err != nil {
		return err
	}
	ch, err := i.ensureCanManageMembers(ctx, input.ChannelID, input.OperatorID)
	if err != nil {
		return err
	}

	target, err := i.workspaceRepo.FindMember(ctx, ch.WorkspaceID, input.TargetUserID)
	if err != nil {
		return fmt.Errorf("failed to verify target user workspace membership: %w", err)
	}
	if target == nil {
		return domerr.ErrUserNotFound
	}

	if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{
		ChannelID: input.ChannelID,
		UserID:    input.TargetUserID,
		Role:      role,
		JoinedAt:  time.Now(),
	}); err != nil {
		return err
	}
	i.recordSystemMessage(ctx, ch, entity.SystemMessageKindMemberAdded, input.OperatorID, input.TargetUserID)
	return nil
}

// JoinPublicChannel は既に参加していても成功します
func (i *channelMemberInteractor) JoinPublicChannel(ctx context.Context, input JoinChannelInput) error {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if ch == nil {
		return domerr.ErrChannelNotFound
	}
	if ch.IsPrivate() {
		return ErrChannelNotPublic
	}
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, ch.ID, input.UserID); err != nil {
		return err
	}

	err = i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{
		ChannelID: input.ChannelID,
		UserID:    input.UserID,
		Role:      entity.ChannelRoleMember,
		JoinedAt:  time.Now(),
	})
	if errors.Is(err, domerr.ErrAlreadyMember) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	i.recordSystemMessage(ctx, ch, entity.SystemMessageKindMemberJoined, input.UserID, input.UserID)
	return nil
}

func (i *channelMemberInteractor) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) error {
	role, err := parseRole(input.Role)
	if err != nil {
		return err
	}
	if _, err := i.ensureCanManageMembers(ctx, input.ChannelID, input.OperatorID); err != nil {
		return err
	}
	return i.changeMember(ctx, input.ChannelID, input.TargetUserID, &role, func(ctx context.Context) error {
		return i.channelMemberRepo.UpdateMemberRole(ctx, input.ChannelID, input.TargetUserID, role)
	})
}

func (i *channelMemberInteractor) RemoveMember(ctx context.Context, input RemoveMemberInput) error {
	ch, err := i.ensureCanManageMembers(ctx, input.ChannelID, input.OperatorID)
	if err != nil {
		return err
	}
	return i.removeMember(ctx, ch, input.TargetUserID)
}

func (i *channelMemberInteractor) LeaveChannel(ctx context.Context, input LeaveChannelInput) error {
	ch, err := i.channelRepo.FindByID(ctx, input.ChannelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if ch == nil {
		return domerr.ErrChannelNotFound
	}
	return i.removeMember(ctx, ch, input.UserID)
}

// removeMember はチャンネルから外し、見られなくなったチャンネルの配信を止めます
func (i *channelMemberInteractor) removeMember(ctx context.Context, ch *entity.Channel, userID string) error {
	err := i.changeMember(ctx, ch.ID, userID, nil, func(ctx context.Context) error {
		return i.channelMemberRepo.RemoveMember(ctx, ch.ID, userID)
	})
	if err != nil {
		return err
	}
	i.revoker.RevokeChannel(ch.WorkspaceID, ch.ID, userID)
	return nil
}
