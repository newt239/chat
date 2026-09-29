package auth

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/audit"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
)

type stubUserRepo struct {
	domainrepository.UserRepository
	users   []*entity.User
	created []*entity.User
}

func (r *stubUserRepo) find(match func(*entity.User) bool) (*entity.User, error) {
	for _, u := range r.users {
		if match(u) {
			return u, nil
		}
	}
	return nil, nil
}

func (r *stubUserRepo) FindByEmail(_ context.Context, email string) (*entity.User, error) {
	return r.find(func(u *entity.User) bool { return u.Email == email })
}

func (r *stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return r.find(func(u *entity.User) bool { return u.ID == id })
}

func (r *stubUserRepo) FindByGoogleSub(_ context.Context, sub string) (*entity.User, error) {
	return r.find(func(u *entity.User) bool { return u.GoogleSub != nil && *u.GoogleSub == sub })
}

func (r *stubUserRepo) Create(_ context.Context, u *entity.User) error {
	u.ID = "new-user"
	r.users = append(r.users, u)
	r.created = append(r.created, u)
	return nil
}

func (r *stubUserRepo) Update(context.Context, *entity.User) error { return nil }

type stubSessionRepo struct {
	domainrepository.SessionRepository
	created []*entity.Session
	active  []*entity.Session
	rotated []string
}

func (r *stubSessionRepo) Create(_ context.Context, s *entity.Session) error {
	r.created = append(r.created, s)
	return nil
}

func (r *stubSessionRepo) FindActiveByTokenHash(_ context.Context, hash string) (*entity.Session, error) {
	for _, s := range r.active {
		if s.RefreshTokenHash == hash {
			return s, nil
		}
	}
	return nil, nil
}

func (r *stubSessionRepo) Rotate(_ context.Context, id string, _ string, _ time.Time) error {
	r.rotated = append(r.rotated, id)
	return nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	added []*entity.WorkspaceMember
}

func (stubWorkspaceRepo) FindByUserID(context.Context, string) ([]*entity.Workspace, error) {
	return []*entity.Workspace{{ID: "ws1"}, {ID: "ws2"}}, nil
}

func (stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	for _, ws := range []*entity.Workspace{
		{ID: "open", SignupEnabled: true, EmailSignupEnabled: true},
		{ID: "google-only", SignupEnabled: true},
		{ID: "closed"},
	} {
		if ws.ID == id {
			return ws, nil
		}
	}
	return nil, nil
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	for _, m := range r.added {
		if m.WorkspaceID == workspaceID && m.UserID == userID {
			return m, nil
		}
	}
	return nil, nil
}

func (r *stubWorkspaceRepo) AddMember(_ context.Context, m *entity.WorkspaceMember) error {
	r.added = append(r.added, m)
	return nil
}

type stubInvitationRepo struct {
	domainrepository.InvitationRepository
	invitations []*entity.Invitation
	accepted    []string
}

func (r *stubInvitationRepo) FindByTokenHash(_ context.Context, hash string) (*entity.Invitation, error) {
	for _, inv := range r.invitations {
		if inv.TokenHash == hash {
			return inv, nil
		}
	}
	return nil, nil
}

func (r *stubInvitationRepo) FindPendingByEmail(_ context.Context, email string, now time.Time) ([]*entity.Invitation, error) {
	var result []*entity.Invitation
	for _, inv := range r.invitations {
		if inv.Email == email && inv.IsPending(now) {
			result = append(result, inv)
		}
	}
	return result, nil
}

func (r *stubInvitationRepo) MarkAccepted(_ context.Context, id string, _ time.Time) error {
	r.accepted = append(r.accepted, id)
	return nil
}

type stubJWT struct{}

func (stubJWT) GenerateToken(userID string, _ time.Duration) (string, error) {
	return "token-" + userID, nil
}
func (stubJWT) VerifyToken(string) (*TokenClaims, error) { return &TokenClaims{UserID: "alice"}, nil }

// ハッシュ化せずにそのまま比較する
type stubPassword struct{}

func (stubPassword) HashPassword(p string) (string, error) { return p, nil }
func (stubPassword) VerifyPassword(p, hash string) error {
	if p != hash {
		return errors.New("mismatch")
	}
	return nil
}

// ID トークンの文字列をそのまま本人情報として扱う
type stubGoogle map[string]*GoogleIdentity

func (g stubGoogle) Verify(_ context.Context, token string) (*GoogleIdentity, error) {
	if identity, ok := g[token]; ok {
		return identity, nil
	}
	return nil, ErrInvalidToken
}

type stubTx struct{}

func (stubTx) Do(ctx context.Context, fn func(ctx context.Context) error) error { return fn(ctx) }

type fixture struct {
	uc          AuthUseCase
	users       *stubUserRepo
	sessions    *stubSessionRepo
	workspaces  *stubWorkspaceRepo
	invitations *stubInvitationRepo
	recorder    *audittest.Recorder
}

var google = stubGoogle{
	"alice":    {Sub: "sub-alice", Email: "alice@example.com", EmailVerified: true},
	"linked":   {Sub: "sub-linked", Email: "someone@example.com", EmailVerified: true},
	"invited":  {Sub: "sub-new", Email: "New@Example.com", EmailVerified: true, Name: "New User"},
	"stranger": {Sub: "sub-stranger", Email: "stranger@example.com", EmailVerified: true},
	"bob":      {Sub: "sub-bob", Email: "bob@example.com", EmailVerified: false},
}

func newFixture(passwordAuthEnabled bool) fixture {
	linkedSub := "sub-linked"
	f := fixture{
		users: &stubUserRepo{users: []*entity.User{
			{ID: "alice", Email: "alice@example.com", PasswordHash: "password123"},
			{ID: "linked", Email: "changed@example.com", PasswordHash: entity.UnusablePasswordHash, GoogleSub: &linkedSub},
		}},
		sessions:   &stubSessionRepo{},
		workspaces: &stubWorkspaceRepo{},
		invitations: &stubInvitationRepo{invitations: []*entity.Invitation{
			{ID: "inv1", WorkspaceID: "ws1", Email: "new@example.com", Role: entity.WorkspaceRoleMember, TokenHash: entity.HashSecretToken("invite-token"), ExpiresAt: time.Now().Add(time.Hour)},
			{ID: "inv2", WorkspaceID: "ws2", Email: "new@example.com", Role: entity.WorkspaceRoleGuest, TokenHash: entity.HashSecretToken("other-token"), ExpiresAt: time.Now().Add(time.Hour)},
			{ID: "expired", WorkspaceID: "ws3", Email: "new@example.com", Role: entity.WorkspaceRoleMember, TokenHash: entity.HashSecretToken("expired-token"), ExpiresAt: time.Now().Add(-time.Hour)},
		}},
		recorder: &audittest.Recorder{},
	}
	settings := Settings{AccessTokenTTL: time.Minute, RefreshTokenTTL: time.Hour, PasswordAuthEnabled: passwordAuthEnabled}
	f.uc = NewAuthInteractor(f.users, f.sessions, f.workspaces, f.invitations, stubJWT{}, stubPassword{}, google, stubTx{}, f.recorder, settings)
	return f
}

func TestLoginRecordsAuditLogInEveryWorkspace(t *testing.T) {
	f := newFixture(true)
	ctx := audit.WithClientInfo(context.Background(), audit.ClientInfo{IPAddress: "203.0.113.1", UserAgent: "Firefox"})

	if _, err := f.uc.Login(ctx, LoginInput{Email: "alice@example.com", Password: "password123"}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(f.sessions.created) != 1 || f.sessions.created[0].IPAddress != "203.0.113.1" || f.sessions.created[0].UserAgent != "Firefox" {
		t.Errorf("セッションにログイン元の端末が保存されていません: %+v", f.sessions.created)
	}
	if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionLogin, entity.AuditActionLogin}) {
		t.Fatalf("参加している全ワークスペースに記録されていません: %v", f.recorder.Actions())
	}
	if f.recorder.Logs[0].WorkspaceID != "ws1" || *f.recorder.Logs[0].ActorID != "alice" {
		t.Errorf("監査ログの内容が期待と異なります: %+v", f.recorder.Logs[0])
	}
}

func TestLoginFailureRecordsAuditLogWithoutActor(t *testing.T) {
	f := newFixture(true)

	if _, err := f.uc.Login(context.Background(), LoginInput{Email: "alice@example.com", Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("エラーが期待と異なります: %v", err)
	}
	if len(f.sessions.created) != 0 {
		t.Errorf("失敗したログインでセッションが作られました")
	}
	if !slices.Equal(f.recorder.Actions(), []entity.AuditAction{entity.AuditActionLoginFailed, entity.AuditActionLoginFailed}) {
		t.Fatalf("ログインの失敗が記録されていません: %v", f.recorder.Actions())
	}
	if f.recorder.Logs[0].ActorID != nil || f.recorder.Logs[0].TargetID != "alice" {
		t.Errorf("失敗したログインは実行者を空にし対象ユーザーを残すはず: %+v", f.recorder.Logs[0])
	}
}

func TestPasswordAuthCanBeDisabled(t *testing.T) {
	f := newFixture(false)

	if _, err := f.uc.Login(context.Background(), LoginInput{Email: "alice@example.com", Password: "password123"}); !errors.Is(err, ErrPasswordAuthDisabled) {
		t.Errorf("パスワード認証が無効なのにログインできました: %v", err)
	}
	if _, err := f.uc.SignUpWithInvitation(context.Background(), SignUpWithInvitationInput{Token: "invite-token", DisplayName: "New", Password: "password123"}); !errors.Is(err, ErrPasswordAuthDisabled) {
		t.Errorf("パスワード認証が無効なのに招待からパスワードで登録できました: %v", err)
	}
	if len(f.sessions.created) != 0 || len(f.users.created) != 0 {
		t.Errorf("セッションやユーザーが作られました")
	}
}

func TestRefreshRotatesSessionFoundByTokenHash(t *testing.T) {
	f := newFixture(true)
	f.sessions.active = []*entity.Session{
		{ID: "s1", UserID: "alice", RefreshTokenHash: entity.HashSecretToken("other")},
		{ID: "s2", UserID: "alice", RefreshTokenHash: entity.HashSecretToken("refresh")},
	}

	out, err := f.uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if out.User.ID != "alice" || !slices.Equal(f.sessions.rotated, []string{"s2"}) || len(f.sessions.created) != 0 {
		t.Errorf("一致したセッションのトークンだけを差し替えるはず: rotated=%v created=%d", f.sessions.rotated, len(f.sessions.created))
	}
	if _, err := f.uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "unknown"}); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("未知のトークンは拒否するはず: %v", err)
	}
}

func TestLoginWithGoogle(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantUser string
		wantErr  error
	}{
		{name: "sub が紐付いたユーザーはメールアドレスが変わってもログインできる", token: "linked", wantUser: "linked"},
		{name: "未紐付けの既存ユーザーはメールアドレスで紐付く", token: "alice", wantUser: "alice"},
		{name: "招待のない未登録のメールアドレスは拒否する", token: "stranger", wantErr: ErrInvitationRequired},
		{name: "確認されていないメールアドレスは拒否する", token: "bob", wantErr: ErrEmailNotVerified},
		{name: "不正な ID トークンは拒否する", token: "forged", wantErr: ErrInvalidToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(false)
			out, err := f.uc.LoginWithGoogle(context.Background(), LoginWithGoogleInput{IDToken: tt.token})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if len(f.users.created) != 0 || len(f.sessions.created) != 0 {
					t.Errorf("拒否したのにユーザーかセッションが作られました")
				}
				return
			}
			if out.User.ID != tt.wantUser || len(f.users.created) != 0 {
				t.Errorf("ログインしたユーザーが期待と異なります: %+v", out.User)
			}
		})
	}
}

func TestLoginWithGoogleLinksSubToExistingUser(t *testing.T) {
	f := newFixture(false)
	if _, err := f.uc.LoginWithGoogle(context.Background(), LoginWithGoogleInput{IDToken: "alice"}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	alice, _ := f.users.FindByID(context.Background(), "alice")
	if alice.GoogleSub == nil || *alice.GoogleSub != "sub-alice" {
		t.Errorf("Google アカウントが紐付いていません: %+v", alice)
	}
}

func TestLoginWithGoogleCreatesInvitedUser(t *testing.T) {
	f := newFixture(false)

	out, err := f.uc.LoginWithGoogle(context.Background(), LoginWithGoogleInput{IDToken: "invited"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(f.users.created) != 1 {
		t.Fatalf("ユーザーが作られていません")
	}
	user := f.users.created[0]
	if user.Email != "new@example.com" || user.DisplayName != "New User" || *user.GoogleSub != "sub-new" || user.PasswordHash != entity.UnusablePasswordHash {
		t.Errorf("作られたユーザーが期待と異なります: %+v", user)
	}
	assertJoinedInvitedWorkspaces(t, f)
	if out.User.ID != user.ID || len(f.sessions.created) != 1 {
		t.Errorf("作ったユーザーでログインしていません: %+v", out.User)
	}
}

func TestSignUpWithInvitation(t *testing.T) {
	f := newFixture(true)

	if _, err := f.uc.SignUpWithInvitation(context.Background(), SignUpWithInvitationInput{Token: "expired-token", DisplayName: "New", Password: "password123"}); !errors.Is(err, ErrInvitationNotFound) {
		t.Fatalf("期限切れの招待は拒否するはず: %v", err)
	}
	if _, err := f.uc.SignUpWithInvitation(context.Background(), SignUpWithInvitationInput{Token: "invite-token", DisplayName: "New", Password: "password123"}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(f.users.created) != 1 || f.users.created[0].PasswordHash != "password123" || f.users.created[0].Email != "new@example.com" {
		t.Fatalf("パスワードを設定したユーザーが作られていません: %+v", f.users.created)
	}
	assertJoinedInvitedWorkspaces(t, f)
}

// 同じメールアドレスへの期限内の招待はすべて受諾し、期限切れの招待は使わない
func assertJoinedInvitedWorkspaces(t *testing.T, f fixture) {
	t.Helper()
	joined := make([]string, 0, len(f.workspaces.added))
	for _, m := range f.workspaces.added {
		joined = append(joined, string(m.Role)+"@"+m.WorkspaceID)
	}
	if !slices.Equal(joined, []string{"member@ws1", "guest@ws2"}) {
		t.Errorf("招待されたワークスペースに参加していません: %v", joined)
	}
	if !slices.Equal(f.invitations.accepted, []string{"inv1", "inv2"}) {
		t.Errorf("招待が受諾済みになっていません: %v", f.invitations.accepted)
	}
}

func TestLoginWithGoogleCreatesUserFromSignupWorkspace(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		wantErr     error
	}{
		{name: "登録を許可したワークスペースなら招待がなくても作る", workspaceID: "google-only"},
		{name: "登録を許可していないワークスペースは拒否する", workspaceID: "closed", wantErr: ErrSignupDisabled},
		{name: "存在しないワークスペースは拒否する", workspaceID: "missing", wantErr: ErrSignupDisabled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(false)
			_, err := f.uc.LoginWithGoogle(context.Background(), LoginWithGoogleInput{IDToken: "stranger", WorkspaceID: &tt.workspaceID})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if len(f.users.created) != 0 || len(f.workspaces.added) != 0 {
					t.Errorf("拒否したのにユーザーかメンバーが作られました")
				}
				return
			}
			assertJoinedAsMember(t, f, tt.workspaceID)
		})
	}
}

func TestSignUp(t *testing.T) {
	tests := []struct {
		name                string
		passwordAuthEnabled bool
		workspaceID         string
		email               string
		wantErr             error
	}{
		{name: "登録とメールでの登録を許可したワークスペースなら作る", passwordAuthEnabled: true, workspaceID: "open", email: "New@Example.com"},
		{name: "メールでの登録を許可していなければ拒否する", passwordAuthEnabled: true, workspaceID: "google-only", email: "new@example.com", wantErr: ErrSignupDisabled},
		{name: "登録を許可していなければ拒否する", passwordAuthEnabled: true, workspaceID: "closed", email: "new@example.com", wantErr: ErrSignupDisabled},
		{name: "パスワード認証が無効なら拒否する", passwordAuthEnabled: false, workspaceID: "open", email: "new@example.com", wantErr: ErrPasswordAuthDisabled},
		{name: "登録済みのメールアドレスは拒否する", passwordAuthEnabled: true, workspaceID: "open", email: "alice@example.com", wantErr: ErrUserAlreadyExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(tt.passwordAuthEnabled)
			_, err := f.uc.SignUp(context.Background(), SignUpInput{WorkspaceID: tt.workspaceID, Email: tt.email, DisplayName: "New", Password: "password123"})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if len(f.users.created) != 0 || len(f.workspaces.added) != 0 || len(f.sessions.created) != 0 {
					t.Errorf("拒否したのにユーザーかメンバーかセッションが作られました")
				}
				return
			}
			if f.users.created[0].Email != "new@example.com" || f.users.created[0].PasswordHash != "password123" {
				t.Errorf("作られたユーザーが期待と異なります: %+v", f.users.created[0])
			}
			assertJoinedAsMember(t, f, tt.workspaceID)
		})
	}
}

func assertJoinedAsMember(t *testing.T, f fixture, workspaceID string) {
	t.Helper()
	if len(f.users.created) != 1 || len(f.sessions.created) != 1 {
		t.Fatalf("ユーザーが作られていないかログインしていません")
	}
	if len(f.workspaces.added) != 1 || f.workspaces.added[0].WorkspaceID != workspaceID || f.workspaces.added[0].Role != entity.WorkspaceRoleMember {
		t.Errorf("ワークスペースにメンバーとして参加していません: %+v", f.workspaces.added)
	}
}

func TestLoginWithGoogleJoinsExistingUserToSignupWorkspace(t *testing.T) {
	f := newFixture(false)
	workspaceID := "google-only"

	for range 2 {
		if _, err := f.uc.LoginWithGoogle(context.Background(), LoginWithGoogleInput{IDToken: "alice", WorkspaceID: &workspaceID}); err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
	}
	if len(f.users.created) != 0 || len(f.workspaces.added) != 1 || f.workspaces.added[0].UserID != "alice" || f.workspaces.added[0].WorkspaceID != workspaceID {
		t.Errorf("既存ユーザーが一度だけメンバーとして参加するはず: %+v", f.workspaces.added)
	}
}
