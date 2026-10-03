// Package admin はワークスペースの管理画面（監査ログ・メンバー管理・権限設定）のユースケースです
package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit"
)

var (
	ErrMemberNotFound       = domerr.New(domerr.ErrNotFound, "メンバーが見つかりません")
	ErrCannotSuspendOwner   = domerr.New(domerr.ErrFailedPrecondition, "オーナーは停止できません")
	ErrCannotSuspendSelf    = domerr.New(domerr.ErrFailedPrecondition, "自分自身は停止できません")
	ErrInvalidPermission    = domerr.New(domerr.ErrValidation, "無効な権限の指定です")
	ErrOwnerOnlyPermissions = domerr.New(domerr.ErrUnauthorized, "管理者の権限はオーナーだけが変更できます")
)

// activityPeriod は管理画面に出すメンバーの投稿数を数える期間です
const activityPeriod = 30 * 24 * time.Hour

// MemberCloser は停止したメンバーのリアルタイム接続を切ります
type MemberCloser interface {
	CloseWorkspaceUser(workspaceID, userID string)
}

type Interactor struct {
	workspaceRepo  domainrepository.WorkspaceRepository
	userRepo       domainrepository.UserRepository
	sessionRepo    domainrepository.SessionRepository
	auditLogRepo   domainrepository.AuditLogRepository
	permissionRepo domainrepository.PermissionRepository
	permissionSvc  domainservice.PermissionService
	recorder       audit.Recorder
	memberCloser   MemberCloser
}

func New(
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	sessionRepo domainrepository.SessionRepository,
	auditLogRepo domainrepository.AuditLogRepository,
	permissionRepo domainrepository.PermissionRepository,
	permissionSvc domainservice.PermissionService,
	recorder audit.Recorder,
	memberCloser MemberCloser,
) *Interactor {
	return &Interactor{
		workspaceRepo:  workspaceRepo,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		auditLogRepo:   auditLogRepo,
		permissionRepo: permissionRepo,
		permissionSvc:  permissionSvc,
		recorder:       recorder,
		memberCloser:   memberCloser,
	}
}

func (i *Interactor) ListMembers(ctx context.Context, input WorkspaceInput) ([]MemberOutput, error) {
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.WorkspaceID, input.RequesterID); err != nil {
		return nil, err
	}
	members, err := i.workspaceRepo.FindMembersByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	users, err := i.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	sessions, err := i.sessionRepo.FindLatestByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load sessions: %w", err)
	}

	activities, err := i.workspaceRepo.FindMemberActivities(ctx, input.WorkspaceID, time.Now().Add(-activityPeriod))
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate member activities: %w", err)
	}

	output := make([]MemberOutput, 0, len(members))
	for _, m := range members {
		out := MemberOutput{
			UserID:      m.UserID,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt,
			SuspendedAt: m.SuspendedAt,
			LastLogin:   sessions[m.UserID],
			Activity:    activities[m.UserID],
		}
		if u := users[m.UserID]; u != nil {
			out.Email = u.Email
			out.DisplayName = u.DisplayName
			out.AvatarURL = u.AvatarURL
		}
		output = append(output, out)
	}
	return output, nil
}

func (i *Interactor) SuspendMember(ctx context.Context, input MemberActionInput) error {
	target, err := i.findTarget(ctx, input)
	if err != nil {
		return err
	}
	if input.TargetUserID == input.OperatorID {
		return ErrCannotSuspendSelf
	}
	if target.member.Role == entity.WorkspaceRoleOwner {
		return ErrCannotSuspendOwner
	}

	if err := i.workspaceRepo.SetMemberSuspended(ctx, input.WorkspaceID, input.TargetUserID, new(time.Now())); err != nil {
		return fmt.Errorf("failed to suspend member: %w", err)
	}
	// 他のワークスペースでは使い続けられるようセッションは失効させず、このワークスペースの接続だけを切る
	i.memberCloser.CloseWorkspaceUser(input.WorkspaceID, input.TargetUserID)

	i.recordMemberAction(ctx, input, target.label, entity.AuditActionMemberSuspended)
	return nil
}

func (i *Interactor) ResumeMember(ctx context.Context, input MemberActionInput) error {
	target, err := i.findTarget(ctx, input)
	if err != nil {
		return err
	}
	if err := i.workspaceRepo.SetMemberSuspended(ctx, input.WorkspaceID, input.TargetUserID, nil); err != nil {
		return fmt.Errorf("failed to resume member: %w", err)
	}

	i.recordMemberAction(ctx, input, target.label, entity.AuditActionMemberResumed)
	return nil
}

type memberTarget struct {
	member *entity.WorkspaceMember
	label  string
}

func (i *Interactor) findTarget(ctx context.Context, input MemberActionInput) (*memberTarget, error) {
	if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, input.WorkspaceID, input.OperatorID); err != nil {
		return nil, err
	}
	member, err := i.workspaceRepo.FindMemberIncludingSuspended(ctx, input.WorkspaceID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load target member: %w", err)
	}
	if member == nil {
		return nil, ErrMemberNotFound
	}
	user, err := i.userRepo.FindByID(ctx, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load target user: %w", err)
	}
	label := input.TargetUserID
	if user != nil {
		label = user.DisplayName
	}
	return &memberTarget{member: member, label: label}, nil
}

func (i *Interactor) recordMemberAction(ctx context.Context, input MemberActionInput, label string, action entity.AuditAction) {
	i.recorder.Record(ctx, entity.AuditLog{
		WorkspaceID: input.WorkspaceID,
		ActorID:     &input.OperatorID,
		Action:      action,
		TargetType:  entity.AuditTargetUser,
		TargetID:    input.TargetUserID,
		TargetLabel: label,
	})
}
