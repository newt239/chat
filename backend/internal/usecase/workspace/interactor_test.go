package workspace

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members   map[string]*entity.WorkspaceMember
	workspace *entity.Workspace
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

func (r *stubWorkspaceRepo) FindByID(context.Context, string) (*entity.Workspace, error) {
	return r.workspace, nil
}

func (r *stubWorkspaceRepo) AddMember(_ context.Context, m *entity.WorkspaceMember) error {
	r.members[m.UserID] = m
	return nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: "name-" + id}, nil
}

type fixture struct {
	uc   *Interactor
	repo *stubWorkspaceRepo
}

func newFixture(members map[string]*entity.WorkspaceMember) fixture {
	repo := &stubWorkspaceRepo{members: members}
	return fixture{uc: New(repo, stubUserRepo{}, nil, nil), repo: repo}
}

func member(role entity.WorkspaceRole) *entity.WorkspaceMember {
	return &entity.WorkspaceMember{Role: role}
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
			joinErr := f.uc.JoinPublicWorkspace(context.Background(), "ws", "u1")
			if tt.wantErr {
				if !errors.Is(err, domerr.ErrWorkspaceNotFound) || joinErr == nil || len(f.repo.members) != 0 {
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

func TestSuspendedMemberCannotRejoin(t *testing.T) {
	suspended := member(entity.WorkspaceRoleMember)
	suspended.SuspendedAt = new(time.Now())
	repo := &stubWorkspaceRepo{members: map[string]*entity.WorkspaceMember{"bob": suspended}, workspace: &entity.Workspace{ID: "ws", IsPublic: true}}
	uc := New(repo, stubUserRepo{}, nil, nil)

	if err := uc.JoinPublicWorkspace(context.Background(), "ws", "bob"); !errors.Is(err, domerr.ErrAlreadyMember) {
		t.Fatalf("停止中のメンバーが参加し直せています: %v", err)
	}
	if repo.members["bob"].SuspendedAt == nil {
		t.Error("参加し直しで停止が解けています")
	}
}
