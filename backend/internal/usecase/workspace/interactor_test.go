package workspace

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
)

// ロール変更の検証に必要な最小限のスタブ
type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members    map[string]*entity.WorkspaceMember
	workspace  *entity.Workspace
	updatedTo  entity.WorkspaceRole
	updateCall int
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

func (r *stubWorkspaceRepo) FindByID(context.Context, string) (*entity.Workspace, error) {
	return r.workspace, nil
}

func (r *stubWorkspaceRepo) AddMember(_ context.Context, m *entity.WorkspaceMember) error {
	r.members[m.UserID] = m
	return nil
}

func (r *stubWorkspaceRepo) UpdateMemberRole(_ context.Context, _ string, _ string, role entity.WorkspaceRole) error {
	r.updateCall++
	r.updatedTo = role
	return nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: "name-" + id}, nil
}

type stubPermissionRepo struct {
	domainrepository.PermissionRepository
	overrides []entity.PermissionOverride
}

func (r *stubPermissionRepo) FindOverrides(context.Context, string) ([]entity.PermissionOverride, error) {
	return r.overrides, nil
}

type fixture struct {
	uc       WorkspaceUseCase
	repo     *stubWorkspaceRepo
	recorder *audittest.Recorder
}

func newFixture(members map[string]*entity.WorkspaceMember, overrides ...entity.PermissionOverride) fixture {
	repo := &stubWorkspaceRepo{members: members}
	recorder := &audittest.Recorder{}
	permissionSvc := domainservice.NewPermissionService(repo, &stubPermissionRepo{overrides: overrides})
	return fixture{uc: NewWorkspaceInteractor(repo, stubUserRepo{}, nil, permissionSvc, recorder), repo: repo, recorder: recorder}
}

func newInteractor(members map[string]*entity.WorkspaceMember) (WorkspaceUseCase, *stubWorkspaceRepo) {
	f := newFixture(members)
	return f.uc, f.repo
}

func member(role entity.WorkspaceRole) *entity.WorkspaceMember {
	return &entity.WorkspaceMember{Role: role}
}

func TestUpdateMemberRole(t *testing.T) {
	tests := []struct {
		name      string
		members   map[string]*entity.WorkspaceMember
		input     UpdateMemberRoleInput
		wantErr   error
		wantCalls int
	}{
		{
			name: "admin は owner を降格できない",
			members: map[string]*entity.WorkspaceMember{
				"admin": member(entity.WorkspaceRoleAdmin),
				"owner": member(entity.WorkspaceRoleOwner),
			},
			input:   UpdateMemberRoleInput{UpdaterID: "admin", UserID: "owner", Role: "member"},
			wantErr: ErrCannotChangeOwnerRole,
		},
		{
			name: "admin は他人を owner に昇格できない",
			members: map[string]*entity.WorkspaceMember{
				"admin":  member(entity.WorkspaceRoleAdmin),
				"member": member(entity.WorkspaceRoleMember),
			},
			input:   UpdateMemberRoleInput{UpdaterID: "admin", UserID: "member", Role: "owner"},
			wantErr: ErrCannotChangeOwnerRole,
		},
		{
			name: "自分自身のロールは変更できない",
			members: map[string]*entity.WorkspaceMember{
				"admin": member(entity.WorkspaceRoleAdmin),
			},
			input:   UpdateMemberRoleInput{UpdaterID: "admin", UserID: "admin", Role: "owner"},
			wantErr: ErrCannotChangeOwnerRole,
		},
		{
			name: "member はロールを変更できない",
			members: map[string]*entity.WorkspaceMember{
				"user":   member(entity.WorkspaceRoleMember),
				"target": member(entity.WorkspaceRoleMember),
			},
			input:   UpdateMemberRoleInput{UpdaterID: "user", UserID: "target", Role: "admin"},
			wantErr: ErrUnauthorized,
		},
		{
			name: "admin は member を admin に昇格できる",
			members: map[string]*entity.WorkspaceMember{
				"admin":  member(entity.WorkspaceRoleAdmin),
				"target": member(entity.WorkspaceRoleMember),
			},
			input:     UpdateMemberRoleInput{UpdaterID: "admin", UserID: "target", Role: "admin"},
			wantCalls: 1,
		},
		{
			name: "owner は他人を owner に昇格できる",
			members: map[string]*entity.WorkspaceMember{
				"owner":  member(entity.WorkspaceRoleOwner),
				"target": member(entity.WorkspaceRoleMember),
			},
			input:     UpdateMemberRoleInput{UpdaterID: "owner", UserID: "target", Role: "owner"},
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo := newInteractor(tt.members)
			_, err := uc.UpdateMemberRole(context.Background(), tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if repo.updateCall != tt.wantCalls {
				t.Errorf("更新回数が期待と異なります: got=%d want=%d", repo.updateCall, tt.wantCalls)
			}
		})
	}
}

func TestUpdateMemberRoleRecordsAuditLog(t *testing.T) {
	f := newFixture(map[string]*entity.WorkspaceMember{
		"admin":  member(entity.WorkspaceRoleAdmin),
		"target": member(entity.WorkspaceRoleMember),
	})
	_, err := f.uc.UpdateMemberRole(context.Background(), UpdateMemberRoleInput{WorkspaceID: "ws", UpdaterID: "admin", UserID: "target", Role: "admin"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(f.recorder.Logs) != 1 {
		t.Fatalf("監査ログが 1 件記録されるはず: got=%d", len(f.recorder.Logs))
	}
	log := f.recorder.Logs[0]
	if log.Action != entity.AuditActionMemberRoleChanged || *log.ActorID != "admin" || log.TargetID != "target" {
		t.Errorf("監査ログの内容が期待と異なります: %+v", log)
	}
	if log.Metadata["from"] != "member" || log.Metadata["to"] != "admin" {
		t.Errorf("変更前後のロールが記録されていません: %+v", log.Metadata)
	}
}

func TestSignupEnabledWorkspace(t *testing.T) {
	tests := []struct {
		name      string
		workspace *entity.Workspace
		wantErr   bool
	}{
		{name: "登録を許可していれば非公開でも参加リンクの情報を返し参加できる", workspace: &entity.Workspace{ID: "ws", Name: "WS", SignupEnabled: true, EmailSignupEnabled: true}},
		{name: "登録を許可していなければ存在しないものとして扱い参加できない", workspace: &entity.Workspace{ID: "ws"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(map[string]*entity.WorkspaceMember{})
			f.repo.workspace = tt.workspace

			info, err := f.uc.GetSignupInfo(context.Background(), "ws")
			_, joinErr := f.uc.JoinPublicWorkspace(context.Background(), JoinPublicWorkspaceInput{WorkspaceID: "ws", UserID: "u1"})
			if tt.wantErr {
				if !errors.Is(err, ErrWorkspaceNotFound) || joinErr == nil || len(f.repo.members) != 0 {
					t.Errorf("登録を許可していないのに情報を返したか参加できました: err=%v joinErr=%v", err, joinErr)
				}
				return
			}
			if err != nil || joinErr != nil {
				t.Fatalf("予期しないエラー: %v %v", err, joinErr)
			}
			if info.Name != "WS" || !info.EmailSignupEnabled || f.repo.members["u1"].Role != entity.WorkspaceRoleMember {
				t.Errorf("返した情報か参加したロールが期待と異なります: %+v %+v", info, f.repo.members["u1"])
			}
		})
	}
}
