package invitation

import (
	"context"
	"errors"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainerrors "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

var (
	ErrAlreadyMember = errors.New("このユーザーは既にワークスペースに参加しています")
	ErrInvalidRole   = errors.New("招待できないロールです")
)

// Sender は招待メールの送信先です。当面はリンクを管理画面でコピーするため送信しない実装を使います
type Sender interface {
	SendInvitation(ctx context.Context, invitation *entity.Invitation, token string) error
}

type CreateInput struct {
	WorkspaceID string
	Email       string
	Role        entity.WorkspaceRole
	RequestedBy string
}

type CreateOutput struct {
	AddedDirectly bool
	Invitation    *InvitationOutput
	Token         string
}

type InvitationOutput struct {
	*entity.Invitation
	InvitedByName string
}

type PreviewOutput struct {
	WorkspaceName string
	Email         string
}

type Interactor struct {
	invitationRepo domainrepository.InvitationRepository
	workspaceRepo  domainrepository.WorkspaceRepository
	userRepo       domainrepository.UserRepository
	permissionSvc  domainservice.PermissionService
	sender         Sender
}

func NewInteractor(
	invitationRepo domainrepository.InvitationRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	userRepo domainrepository.UserRepository,
	permissionSvc domainservice.PermissionService,
	sender Sender,
) *Interactor {
	return &Interactor{
		invitationRepo: invitationRepo,
		workspaceRepo:  workspaceRepo,
		userRepo:       userRepo,
		permissionSvc:  permissionSvc,
		sender:         sender,
	}
}

// Create は既存ユーザーなら直ちにワークスペースへ追加し、未登録なら招待を作ります。管理者として招待できるのは管理者だけです
func (i *Interactor) Create(ctx context.Context, input CreateInput) (*CreateOutput, error) {
	requester, err := i.permissionSvc.Ensure(ctx, input.WorkspaceID, input.RequestedBy, entity.PermissionInviteMembers)
	if err != nil {
		return nil, err
	}
	switch input.Role {
	case entity.WorkspaceRoleAdmin:
		if !requester.IsAdmin() {
			return nil, domainerrors.ErrUnauthorized
		}
	case entity.WorkspaceRoleMember, entity.WorkspaceRoleGuest:
	default:
		return nil, ErrInvalidRole
	}

	email := entity.NormalizeEmail(input.Email)
	user, err := i.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user != nil && !user.IsBot {
		return i.addExistingUser(ctx, input, user)
	}

	token, hash, err := entity.NewSecretToken()
	if err != nil {
		return nil, err
	}
	invitation := &entity.Invitation{
		WorkspaceID: input.WorkspaceID,
		Email:       email,
		Role:        input.Role,
		TokenHash:   hash,
		InvitedBy:   input.RequestedBy,
		ExpiresAt:   time.Now().Add(entity.InvitationTTL),
	}
	if err := i.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}
	if err := i.sender.SendInvitation(ctx, invitation, token); err != nil {
		return nil, err
	}
	inviter, err := i.userRepo.FindByID(ctx, input.RequestedBy)
	if err != nil {
		return nil, err
	}
	out := &InvitationOutput{Invitation: invitation}
	if inviter != nil {
		out.InvitedByName = inviter.DisplayName
	}
	return &CreateOutput{Invitation: out, Token: token}, nil
}

func (i *Interactor) addExistingUser(ctx context.Context, input CreateInput, user *entity.User) (*CreateOutput, error) {
	existing, err := i.workspaceRepo.FindMember(ctx, input.WorkspaceID, user.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyMember
	}
	member := &entity.WorkspaceMember{WorkspaceID: input.WorkspaceID, UserID: user.ID, Role: input.Role, JoinedAt: time.Now()}
	if err := i.workspaceRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}
	return &CreateOutput{AddedDirectly: true}, nil
}

func (i *Interactor) List(ctx context.Context, workspaceID, requestedBy string) ([]InvitationOutput, error) {
	if _, err := i.permissionSvc.Ensure(ctx, workspaceID, requestedBy, entity.PermissionInviteMembers); err != nil {
		return nil, err
	}
	invitations, err := i.invitationRepo.FindPendingByWorkspaceID(ctx, workspaceID, time.Now())
	if err != nil {
		return nil, err
	}
	inviterIDs := make([]string, 0, len(invitations))
	for _, inv := range invitations {
		inviterIDs = append(inviterIDs, inv.InvitedBy)
	}
	inviters, err := i.userRepo.FindByIDs(ctx, inviterIDs)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string, len(inviters))
	for _, u := range inviters {
		names[u.ID] = u.DisplayName
	}
	result := make([]InvitationOutput, 0, len(invitations))
	for _, inv := range invitations {
		result = append(result, InvitationOutput{Invitation: inv, InvitedByName: names[inv.InvitedBy]})
	}
	return result, nil
}

func (i *Interactor) Revoke(ctx context.Context, workspaceID, invitationID, requestedBy string) error {
	if _, err := i.permissionSvc.Ensure(ctx, workspaceID, requestedBy, entity.PermissionInviteMembers); err != nil {
		return err
	}
	return i.invitationRepo.Delete(ctx, workspaceID, invitationID)
}

// Preview は招待リンクを開いた人に招待先を示します。ログインしていなくても呼べます
func (i *Interactor) Preview(ctx context.Context, token string) (*PreviewOutput, error) {
	invitation, err := i.invitationRepo.FindByTokenHash(ctx, entity.HashSecretToken(token))
	if err != nil {
		return nil, err
	}
	if invitation == nil || !invitation.IsPending(time.Now()) {
		return nil, domainerrors.ErrInvitationNotFound
	}
	workspace, err := i.workspaceRepo.FindByID(ctx, invitation.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return &PreviewOutput{WorkspaceName: workspace.Name, Email: invitation.Email}, nil
}
