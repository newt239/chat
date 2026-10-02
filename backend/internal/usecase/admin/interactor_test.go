package admin

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members map[string]*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	m := r.members[userID]
	if m == nil || m.SuspendedAt != nil {
		return nil, nil
	}
	return m, nil
}

func (r *stubWorkspaceRepo) FindMemberIncludingSuspended(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

func (r *stubWorkspaceRepo) SetMemberSuspended(_ context.Context, _ string, userID string, at *time.Time) error {
	r.members[userID].SuspendedAt = at
	return nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: "name-" + id}, nil
}

func (stubUserRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, &entity.User{ID: id, DisplayName: "name-" + id})
	}
	return users, nil
}

type stubSessionRepo struct {
	domainrepository.SessionRepository
	revoked []string
}

func (r *stubSessionRepo) RevokeAllByUserID(_ context.Context, userID string) error {
	r.revoked = append(r.revoked, userID)
	return nil
}

type stubAuditLogRepo struct {
	domainrepository.AuditLogRepository
	logs []*entity.AuditLog
}

func (r *stubAuditLogRepo) List(_ context.Context, _ entity.AuditLogFilter) (*entity.AuditLogPage, error) {
	return &entity.AuditLogPage{Logs: r.logs}, nil
}

type stubPermissionRepo struct {
	domainrepository.PermissionRepository
	overrides []entity.PermissionOverride
}

func (r *stubPermissionRepo) FindOverrides(context.Context, string) ([]entity.PermissionOverride, error) {
	return r.overrides, nil
}

func (r *stubPermissionRepo) Upsert(_ context.Context, _ string, o entity.PermissionOverride) error {
	r.overrides = append(r.overrides, o)
	return nil
}

type fixture struct {
	uc          *Interactor
	members     map[string]*entity.WorkspaceMember
	sessions    *stubSessionRepo
	auditLogs   *stubAuditLogRepo
	permissions *stubPermissionRepo
	recorder    *audittest.Recorder
	closer      *stubCloser
}

type stubCloser struct{ closed []string }

func (c *stubCloser) CloseWorkspaceUser(workspaceID, userID string) {
	c.closed = append(c.closed, workspaceID+"/"+userID)
}

func newFixture() *fixture {
	f := &fixture{
		members: map[string]*entity.WorkspaceMember{
			"owner":  {UserID: "owner", Role: entity.WorkspaceRoleOwner},
			"admin":  {UserID: "admin", Role: entity.WorkspaceRoleAdmin},
			"member": {UserID: "member", Role: entity.WorkspaceRoleMember},
		},
		sessions:    &stubSessionRepo{},
		auditLogs:   &stubAuditLogRepo{},
		permissions: &stubPermissionRepo{},
		recorder:    &audittest.Recorder{},
		closer:      &stubCloser{},
	}
	workspaceRepo := &stubWorkspaceRepo{members: f.members}
	f.uc = NewInteractor(
		workspaceRepo,
		stubUserRepo{},
		f.sessions,
		f.auditLogs,
		f.permissions,
		nil,
		domainservice.NewPermissionService(workspaceRepo, f.permissions),
		f.recorder,
		f.closer,
	)
	return f
}

func TestSuspendMember(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		target   string
		wantErr  error
	}{
		{name: "管理者はメンバーを停止できる", operator: "admin", target: "member"},
		{name: "メンバーは停止できない", operator: "member", target: "admin", wantErr: domerr.ErrUnauthorized},
		{name: "オーナーは停止できない", operator: "admin", target: "owner", wantErr: ErrCannotSuspendOwner},
		{name: "自分自身は停止できない", operator: "admin", target: "admin", wantErr: ErrCannotSuspendSelf},
		{name: "存在しないメンバーは停止できない", operator: "admin", target: "nobody", wantErr: ErrMemberNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			err := f.uc.SuspendMember(context.Background(), MemberActionInput{WorkspaceID: "ws", TargetUserID: tt.target, OperatorID: tt.operator})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if len(f.closer.closed) != 0 || len(f.recorder.Logs) != 0 {
					t.Errorf("失敗時に接続の切断や記録が行われました")
				}
				return
			}
			if f.members[tt.target].SuspendedAt == nil {
				t.Errorf("停止状態になっていません")
			}
			if len(f.sessions.revoked) != 0 {
				t.Errorf("他のワークスペースで使えるようセッションは失効させないはず: %v", f.sessions.revoked)
			}
			if !slices.Equal(f.closer.closed, []string{"ws/" + tt.target}) {
				t.Errorf("このワークスペースの接続が切られていません: %v", f.closer.closed)
			}
			if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionMemberSuspended}) {
				t.Errorf("監査ログが期待と異なります: %v", f.recorder.Actions())
			}
		})
	}
}

func TestSuspendedAdminLosesAccess(t *testing.T) {
	f := newFixture()
	if err := f.uc.SuspendMember(context.Background(), MemberActionInput{WorkspaceID: "ws", TargetUserID: "admin", OperatorID: "owner"}); err != nil {
		t.Fatalf("停止に失敗しました: %v", err)
	}

	_, err := f.uc.ListAuditLogs(context.Background(), ListAuditLogsInput{AuditLogQuery: AuditLogQuery{WorkspaceID: "ws", RequesterID: "admin"}, Limit: 10})
	if !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("停止中の管理者が管理画面を使えてしまいます: %v", err)
	}

	if err := f.uc.ResumeMember(context.Background(), MemberActionInput{WorkspaceID: "ws", TargetUserID: "admin", OperatorID: "owner"}); err != nil {
		t.Fatalf("再開に失敗しました: %v", err)
	}
	if f.members["admin"].SuspendedAt != nil {
		t.Errorf("再開後も停止状態のままです")
	}
	if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionMemberSuspended, entity.AuditActionMemberResumed}) {
		t.Errorf("監査ログが期待と異なります: %v", f.recorder.Actions())
	}
}

func TestListAuditLogsRequiresAdmin(t *testing.T) {
	f := newFixture()
	_, err := f.uc.ListAuditLogs(context.Background(), ListAuditLogsInput{AuditLogQuery: AuditLogQuery{WorkspaceID: "ws", RequesterID: "member"}, Limit: 10})
	if !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("メンバーが監査ログを閲覧できてしまいます: %v", err)
	}
}

func TestExportAuditLogs(t *testing.T) {
	actor := "member"
	f := newFixture()
	f.auditLogs.logs = []*entity.AuditLog{{
		ActorID:     &actor,
		Action:      entity.AuditActionChannelCreated,
		TargetType:  entity.AuditTargetChannel,
		TargetLabel: "=HYPERLINK(\"x\")",
		Metadata:    map[string]string{"private": "false"},
		CreatedAt:   time.Date(2026, 9, 28, 1, 2, 3, 0, time.UTC),
	}}

	out, err := f.uc.ExportAuditLogs(context.Background(), AuditLogQuery{WorkspaceID: "ws", RequesterID: "admin"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out.Content), "\n")
	if len(lines) != 2 {
		t.Fatalf("ヘッダーと 1 行が出力されるはず: %q", out.Content)
	}
	if !strings.Contains(lines[1], "2026-09-28T01:02:03Z,member,name-member,channel_created") {
		t.Errorf("行の内容が期待と異なります: %s", lines[1])
	}
	if !strings.Contains(lines[1], `"'=HYPERLINK(""x"")"`) {
		t.Errorf("数式がエスケープされていません: %s", lines[1])
	}
	if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionAuditLogExported}) {
		t.Errorf("エクスポートが監査ログに残っていません: %v", f.recorder.Actions())
	}

	if _, err := f.uc.ExportAuditLogs(context.Background(), AuditLogQuery{WorkspaceID: "ws", RequesterID: "member"}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Errorf("メンバーが監査ログを書き出せてしまいます: %v", err)
	}
}

func TestUpdatePermission(t *testing.T) {
	tests := []struct {
		name       string
		operator   string
		role       entity.WorkspaceRole
		allowed    bool
		wantErr    error
		wantRecord bool
	}{
		{name: "管理者はメンバーの権限を変更できる", operator: "admin", role: entity.WorkspaceRoleMember, allowed: true, wantRecord: true},
		{name: "既定値と同じなら何も記録しない", operator: "admin", role: entity.WorkspaceRoleMember, allowed: false},
		{name: "管理者は管理者の権限を変更できない", operator: "admin", role: entity.WorkspaceRoleAdmin, allowed: false, wantErr: ErrOwnerOnlyPermissions},
		{name: "オーナーは管理者の権限を変更できる", operator: "owner", role: entity.WorkspaceRoleAdmin, allowed: false, wantRecord: true},
		{name: "メンバーは権限を変更できない", operator: "member", role: entity.WorkspaceRoleMember, allowed: true, wantErr: domerr.ErrUnauthorized},
		{name: "オーナーの権限は設定できない", operator: "owner", role: entity.WorkspaceRoleOwner, allowed: false, wantErr: ErrInvalidPermission},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			err := f.uc.UpdatePermission(context.Background(), UpdatePermissionInput{
				WorkspaceID: "ws", OperatorID: tt.operator, Role: tt.role, Permission: entity.PermissionDeleteOthersMessages, Allowed: tt.allowed,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if got := len(f.recorder.Logs) == 1; got != tt.wantRecord {
				t.Fatalf("監査ログの記録が期待と異なります: %v", f.recorder.Logs)
			}
			if tt.wantRecord && f.recorder.Logs[0].Metadata["permission"] != string(entity.PermissionDeleteOthersMessages) {
				t.Errorf("変更した権限が記録されていません: %v", f.recorder.Logs[0].Metadata)
			}
		})
	}
}

func TestGetPermissionsForMember(t *testing.T) {
	f := newFixture()
	out, err := f.uc.GetPermissions(context.Background(), WorkspaceInput{WorkspaceID: "ws", RequesterID: "member"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if out.RequesterRole != entity.WorkspaceRoleMember || out.Matrix.Allows(out.RequesterRole, entity.PermissionDeleteOthersMessages) {
		t.Errorf("既定の権限が期待と異なります: %+v", out)
	}
}
