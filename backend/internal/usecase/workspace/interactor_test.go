package workspace

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

// ロール変更の検証に必要な最小限のスタブ
type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members    map[string]*entity.WorkspaceMember
	updatedTo  entity.WorkspaceRole
	updateCall int
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

func (r *stubWorkspaceRepo) UpdateMemberRole(_ context.Context, _ string, _ string, role entity.WorkspaceRole) error {
	r.updateCall++
	r.updatedTo = role
	return nil
}

func newInteractor(members map[string]*entity.WorkspaceMember) (WorkspaceUseCase, *stubWorkspaceRepo) {
	repo := &stubWorkspaceRepo{members: members}
	return NewWorkspaceInteractor(repo, nil), repo
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
