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

var alice = &entity.User{ID: "alice", Email: "alice@example.com", PasswordHash: "password123"}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByEmail(context.Context, string) (*entity.User, error) {
	return alice, nil
}

func (stubUserRepo) FindByID(context.Context, string) (*entity.User, error) {
	return alice, nil
}

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

func (r *stubSessionRepo) FindActiveByUserID(context.Context, string) ([]*entity.Session, error) {
	return r.active, nil
}

func (r *stubSessionRepo) Rotate(_ context.Context, id string, _ string, _ time.Time) error {
	r.rotated = append(r.rotated, id)
	return nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindByUserID(context.Context, string) ([]*entity.Workspace, error) {
	return []*entity.Workspace{{ID: "ws1"}, {ID: "ws2"}}, nil
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

func newInteractor() (AuthUseCase, *stubSessionRepo, *audittest.Recorder) {
	sessions := &stubSessionRepo{}
	recorder := &audittest.Recorder{}
	return NewAuthInteractor(stubUserRepo{}, sessions, stubWorkspaceRepo{}, stubJWT{}, stubPassword{}, recorder), sessions, recorder
}

func TestLoginRecordsAuditLogInEveryWorkspace(t *testing.T) {
	uc, sessions, recorder := newInteractor()
	ctx := audit.WithClientInfo(context.Background(), audit.ClientInfo{IPAddress: "203.0.113.1", UserAgent: "Firefox"})

	if _, err := uc.Login(ctx, LoginInput{Email: alice.Email, Password: "password123"}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(sessions.created) != 1 || sessions.created[0].IPAddress != "203.0.113.1" || sessions.created[0].UserAgent != "Firefox" {
		t.Errorf("セッションにログイン元の端末が保存されていません: %+v", sessions.created)
	}
	if !slices.Equal(recorder.Actions(), []entity.AuditAction{entity.AuditActionLogin, entity.AuditActionLogin}) {
		t.Fatalf("参加している全ワークスペースに記録されていません: %v", recorder.Actions())
	}
	if recorder.Logs[0].WorkspaceID != "ws1" || *recorder.Logs[0].ActorID != "alice" {
		t.Errorf("監査ログの内容が期待と異なります: %+v", recorder.Logs[0])
	}
}

func TestLoginFailureRecordsAuditLogWithoutActor(t *testing.T) {
	uc, sessions, recorder := newInteractor()

	if _, err := uc.Login(context.Background(), LoginInput{Email: alice.Email, Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("エラーが期待と異なります: %v", err)
	}
	if len(sessions.created) != 0 {
		t.Errorf("失敗したログインでセッションが作られました")
	}
	if !slices.Equal(recorder.Actions(), []entity.AuditAction{entity.AuditActionLoginFailed, entity.AuditActionLoginFailed}) {
		t.Fatalf("ログインの失敗が記録されていません: %v", recorder.Actions())
	}
	if recorder.Logs[0].ActorID != nil || recorder.Logs[0].TargetID != "alice" {
		t.Errorf("失敗したログインは実行者を空にし対象ユーザーを残すはず: %+v", recorder.Logs[0])
	}
}

func TestRefreshRotatesExistingSession(t *testing.T) {
	uc, sessions, _ := newInteractor()
	sessions.active = []*entity.Session{{ID: "s1", RefreshTokenHash: "other"}, {ID: "s2", RefreshTokenHash: "refresh"}}

	if _, err := uc.RefreshToken(context.Background(), RefreshTokenInput{RefreshToken: "refresh"}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if !slices.Equal(sessions.rotated, []string{"s2"}) || len(sessions.created) != 0 {
		t.Errorf("一致したセッションのトークンだけを差し替えるはず: rotated=%v created=%d", sessions.rotated, len(sessions.created))
	}
}
