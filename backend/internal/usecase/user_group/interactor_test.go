package user_group

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
	members map[string]*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	return &entity.Workspace{ID: id}, nil
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

type stubUserGroupRepo struct {
	domainrepository.UserGroupRepository
}

func (stubUserGroupRepo) FindByName(context.Context, string, string) (*entity.UserGroup, error) {
	return nil, nil
}

func (stubUserGroupRepo) Create(context.Context, *entity.UserGroup) error {
	return nil
}

func TestCreateUserGroupRequiresAdmin(t *testing.T) {
	workspaceRepo := &stubWorkspaceRepo{members: map[string]*entity.WorkspaceMember{
		"owner":  {UserID: "owner", Role: entity.WorkspaceRoleOwner},
		"admin":  {UserID: "admin", Role: entity.WorkspaceRoleAdmin},
		"member": {UserID: "member", Role: entity.WorkspaceRoleMember},
	}}
	uc := NewUserGroupInteractor(stubUserGroupRepo{}, workspaceRepo, nil)

	tests := []struct {
		userID  string
		wantErr error
	}{
		{userID: "owner"},
		{userID: "admin"},
		{userID: "member", wantErr: domerr.ErrUnauthorized},
		{userID: "outsider", wantErr: domerr.ErrUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.userID, func(t *testing.T) {
			_, err := uc.CreateUserGroup(context.Background(), CreateUserGroupInput{WorkspaceID: "ws", Name: "dev", CreatedBy: tt.userID})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got=%v want=%v", err, tt.wantErr)
			}
		})
	}
}
