package invitation

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members map[string]*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	if m := r.members[userID]; m != nil && m.SuspendedAt == nil {
		return m, nil
	}
	return nil, nil
}

func (r *stubWorkspaceRepo) FindMemberIncludingSuspended(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

func (r *stubWorkspaceRepo) AddMember(_ context.Context, m *entity.WorkspaceMember) error {
	r.members[m.UserID] = m
	return nil
}

func (r *stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	return &entity.Workspace{ID: id, Name: "Acme"}, nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	if email == "existing@example.com" {
		return &entity.User{ID: "existing", Email: email}, nil
	}
	return nil, nil
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: "name-" + id}, nil
}

type stubInvitationRepo struct {
	domainrepository.InvitationRepository
	created []*entity.Invitation
}

func (r *stubInvitationRepo) Create(_ context.Context, inv *entity.Invitation) error {
	inv.ID = "inv"
	r.created = append(r.created, inv)
	return nil
}

func (r *stubInvitationRepo) FindByTokenHash(_ context.Context, hash string) (*entity.Invitation, error) {
	for _, inv := range r.created {
		if inv.TokenHash == hash {
			return inv, nil
		}
	}
	return nil, nil
}

type stubPermissionRepo struct {
	domainrepository.PermissionRepository
	overrides []entity.PermissionOverride
}

func (r *stubPermissionRepo) FindOverrides(context.Context, string) ([]entity.PermissionOverride, error) {
	return r.overrides, nil
}

type recordingSender struct {
	tokens []string
}

func (s *recordingSender) SendInvitation(_ context.Context, _ *entity.Invitation, token string) error {
	s.tokens = append(s.tokens, token)
	return nil
}

type fixture struct {
	uc          *Interactor
	workspaces  *stubWorkspaceRepo
	invitations *stubInvitationRepo
	sender      *recordingSender
}

func newFixture(role entity.WorkspaceRole, overrides ...entity.PermissionOverride) fixture {
	f := fixture{
		workspaces:  &stubWorkspaceRepo{members: map[string]*entity.WorkspaceMember{"user": {Role: role}}},
		invitations: &stubInvitationRepo{},
		sender:      &recordingSender{},
	}
	permissionSvc := domainservice.NewPermissionService(f.workspaces, &stubPermissionRepo{overrides: overrides})
	f.uc = NewInteractor(f.invitations, f.workspaces, stubUserRepo{}, permissionSvc, f.sender)
	return f
}

func TestCreatePermission(t *testing.T) {
	allowMemberInvite := entity.PermissionOverride{Role: entity.WorkspaceRoleMember, Permission: entity.PermissionInviteMembers, Allowed: true}
	tests := []struct {
		name      string
		overrides []entity.PermissionOverride
		role      entity.WorkspaceRole
		wantErr   error
	}{
		{name: "既定ではメンバーは招待できない", role: entity.WorkspaceRoleMember, wantErr: domerr.ErrUnauthorized},
		{name: "権限を許可するとメンバーも招待できる", overrides: []entity.PermissionOverride{allowMemberInvite}, role: entity.WorkspaceRoleMember},
		{name: "招待を許可されたメンバーでも管理者としては招待できない", overrides: []entity.PermissionOverride{allowMemberInvite}, role: entity.WorkspaceRoleAdmin, wantErr: domerr.ErrUnauthorized},
		{name: "オーナーとしては招待できない", overrides: []entity.PermissionOverride{allowMemberInvite}, role: entity.WorkspaceRoleOwner, wantErr: domerr.ErrInvalidRole},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(entity.WorkspaceRoleMember, tt.overrides...)
			_, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: "ws", Email: "new@example.com", Role: tt.role, RequestedBy: "user"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateAddsExistingUserDirectly(t *testing.T) {
	f := newFixture(entity.WorkspaceRoleAdmin)

	out, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: "ws", Email: "Existing@example.com", Role: entity.WorkspaceRoleMember, RequestedBy: "user"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if !out.AddedDirectly || out.Token != "" || len(f.invitations.created) != 0 {
		t.Errorf("既存ユーザーは招待を作らずに追加するはず: %+v", out)
	}
	if m := f.workspaces.members["existing"]; m == nil || m.Role != entity.WorkspaceRoleMember {
		t.Errorf("既存ユーザーがワークスペースに追加されていません: %+v", m)
	}

	if _, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: "ws", Email: "existing@example.com", Role: entity.WorkspaceRoleMember, RequestedBy: "user"}); !errors.Is(err, domerr.ErrAlreadyMember) {
		t.Errorf("参加済みのユーザーは招待できないはず: %v", err)
	}
}

func TestCreateInvitationForNewEmail(t *testing.T) {
	f := newFixture(entity.WorkspaceRoleAdmin)

	out, err := f.uc.Create(context.Background(), CreateInput{WorkspaceID: "ws", Email: " New@Example.com ", Role: entity.WorkspaceRoleGuest, RequestedBy: "user"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if out.AddedDirectly || out.Token == "" || out.Invitation.InvitedByName != "name-user" {
		t.Fatalf("招待の出力が期待と異なります: %+v", out)
	}
	inv := f.invitations.created[0]
	if inv.Email != "new@example.com" || inv.Role != entity.WorkspaceRoleGuest || inv.TokenHash != entity.HashSecretToken(out.Token) {
		t.Errorf("招待はメールアドレスを正規化し、トークンをハッシュで保存するはず: %+v", inv)
	}
	if len(f.sender.tokens) != 1 || f.sender.tokens[0] != out.Token {
		t.Errorf("招待の送信に平文のトークンが渡されていません: %v", f.sender.tokens)
	}

	preview, err := f.uc.Preview(context.Background(), out.Token)
	if err != nil || preview.WorkspaceName != "Acme" || preview.Email != "new@example.com" {
		t.Errorf("招待リンクから招待先を確認できません: %+v %v", preview, err)
	}
	if _, err := f.uc.Preview(context.Background(), "unknown"); !errors.Is(err, domerr.ErrInvitationNotFound) {
		t.Errorf("未知のトークンは見つからないはず: %v", err)
	}
}
