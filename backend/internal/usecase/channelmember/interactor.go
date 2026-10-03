package channelmember

import (
	"context"
	"errors"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrNotMember        = domerr.New(domerr.ErrFailedPrecondition, "ユーザーはメンバーではありません")
	ErrChannelNotPublic = domerr.New(domerr.ErrUnauthorized, "このチャンネルは公開されていません")
	ErrLastAdminRemoval = domerr.New(domerr.ErrFailedPrecondition, "最後の管理者は削除できません")
)

// MemberInput は操作する人が対象者を招待・削除したりロールを変えたりする入力です。Role は削除では使わない
type MemberInput struct {
	ChannelID    string
	OperatorID   string
	TargetUserID string
	Role         entity.ChannelRole
}

type MemberOutput struct {
	User *entity.User
	Role entity.ChannelRole
}

// ChannelRevoker はチャンネルから外れたユーザーへのリアルタイム配信を止めます
type ChannelRevoker interface {
	RevokeChannel(workspaceID, channelID, userID string)
}

type Interactor struct {
	channelRepo       domainrepository.ChannelRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	workspaceRepo     domainrepository.WorkspaceRepository
	userRepo          domainrepository.UserRepository
	systemMessages    *messageuc.SystemMessages
	channelAccessSvc  service.ChannelAccessService
	txManager         domaintransaction.Manager
	revoker           ChannelRevoker
}

func New(
	channelRepo domainrepository.ChannelRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	systemMessages *messageuc.SystemMessages,
	channelAccessSvc service.ChannelAccessService,
	txManager domaintransaction.Manager,
	revoker ChannelRevoker,
) *Interactor {
	return &Interactor{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		workspaceRepo:     workspaceRepo,
		userRepo:          userRepo,
		systemMessages:    systemMessages,
		channelAccessSvc:  channelAccessSvc,
		txManager:         txManager,
		revoker:           revoker,
	}
}

func (i *Interactor) recordSystemMessage(ctx context.Context, ch *entity.Channel, kind entity.SystemMessageKind, actorID, targetUserID string) {
	i.systemMessages.Record(ctx, ch, kind, actorID, map[string]any{"userId": targetUserID})
}

// ensureCanManageMembers は招待・削除・ロール変更を行えるのがワークスペースの管理者とチャンネルの作成者だけであることを確かめます
func (i *Interactor) ensureCanManageMembers(ctx context.Context, channelID, operatorID string) (*entity.Channel, error) {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, operatorID)
	if err != nil || ch.CreatedBy == operatorID {
		return ch, err
	}
	if _, err := service.EnsureAdmin(ctx, i.workspaceRepo, ch.WorkspaceID, operatorID); err != nil {
		return nil, err
	}
	return ch, nil
}

// changeMember は最後の管理者がいなくなる変更を拒否してから change を実行します。newRole が nil なら対象者を外す変更として扱う
func (i *Interactor) changeMember(ctx context.Context, channelID, userID string, newRole *entity.ChannelRole, change func(ctx context.Context) error) error {
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

func validateRole(role entity.ChannelRole) error {
	if role != entity.ChannelRoleMember && role != entity.ChannelRoleAdmin {
		return domerr.ErrInvalidRole
	}
	return nil
}

func (i *Interactor) ListMembers(ctx context.Context, channelID, userID string) ([]MemberOutput, error) {
	if _, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID); err != nil {
		return nil, err
	}
	members, err := i.channelMemberRepo.FindMembersByChannelIDs(ctx, []string{channelID})
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
	outputs := make([]MemberOutput, 0, len(members))
	for _, m := range members {
		if user := users[m.UserID]; user != nil {
			outputs = append(outputs, MemberOutput{User: user, Role: m.Role})
		}
	}
	return outputs, nil
}

// InviteMember は既に参加していれば ErrAlreadyMember を返します
func (i *Interactor) InviteMember(ctx context.Context, input MemberInput) error {
	if err := validateRole(input.Role); err != nil {
		return err
	}
	ch, err := i.ensureCanManageMembers(ctx, input.ChannelID, input.OperatorID)
	if err != nil {
		return err
	}
	if _, err := service.EnsureMember(ctx, i.workspaceRepo, ch.WorkspaceID, input.TargetUserID); err != nil {
		if errors.Is(err, domerr.ErrUnauthorized) {
			return domerr.ErrUserNotFound
		}
		return err
	}
	if err := i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: input.ChannelID, UserID: input.TargetUserID, Role: input.Role}); err != nil {
		return err
	}
	i.recordSystemMessage(ctx, ch, entity.SystemMessageKindMemberAdded, input.OperatorID, input.TargetUserID)
	return nil
}

// JoinPublicChannel は既に参加していても成功します
func (i *Interactor) JoinPublicChannel(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelAccessSvc.EnsureChannelAccess(ctx, channelID, userID)
	if err != nil {
		return err
	}
	if ch.IsPrivate() {
		return ErrChannelNotPublic
	}
	err = i.channelMemberRepo.AddMember(ctx, &entity.ChannelMember{ChannelID: channelID, UserID: userID, Role: entity.ChannelRoleMember})
	if errors.Is(err, domerr.ErrAlreadyMember) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	i.recordSystemMessage(ctx, ch, entity.SystemMessageKindMemberJoined, userID, userID)
	return nil
}

func (i *Interactor) UpdateMemberRole(ctx context.Context, input MemberInput) error {
	if err := validateRole(input.Role); err != nil {
		return err
	}
	if _, err := i.ensureCanManageMembers(ctx, input.ChannelID, input.OperatorID); err != nil {
		return err
	}
	return i.changeMember(ctx, input.ChannelID, input.TargetUserID, &input.Role, func(ctx context.Context) error {
		return i.channelMemberRepo.UpdateMemberRole(ctx, input.ChannelID, input.TargetUserID, input.Role)
	})
}

func (i *Interactor) RemoveMember(ctx context.Context, input MemberInput) error {
	ch, err := i.ensureCanManageMembers(ctx, input.ChannelID, input.OperatorID)
	if err != nil {
		return err
	}
	return i.removeMember(ctx, ch, input.TargetUserID)
}

func (i *Interactor) LeaveChannel(ctx context.Context, channelID, userID string) error {
	ch, err := i.channelRepo.FindByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("failed to find channel: %w", err)
	}
	if ch == nil {
		return domerr.ErrChannelNotFound
	}
	return i.removeMember(ctx, ch, userID)
}

// removeMember はチャンネルから外し、見られなくなったチャンネルの配信を止めます
func (i *Interactor) removeMember(ctx context.Context, ch *entity.Channel, userID string) error {
	err := i.changeMember(ctx, ch.ID, userID, nil, func(ctx context.Context) error {
		return i.channelMemberRepo.RemoveMember(ctx, ch.ID, userID)
	})
	if err != nil {
		return err
	}
	i.revoker.RevokeChannel(ch.WorkspaceID, ch.ID, userID)
	return nil
}
