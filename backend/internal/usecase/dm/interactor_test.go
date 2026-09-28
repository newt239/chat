package dm

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	"github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	repository.WorkspaceRepository
	members map[string]*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return r.members[userID], nil
}

func newInteractor(memberIDs ...string) *Interactor {
	members := make(map[string]*entity.WorkspaceMember, len(memberIDs))
	for _, id := range memberIDs {
		members[id] = &entity.WorkspaceMember{Role: entity.WorkspaceRoleMember}
	}
	return NewInteractor(nil, nil, nil, nil, &stubWorkspaceRepo{members: members})
}

func TestCreateDMRejectsNonWorkspaceMember(t *testing.T) {
	uc := newInteractor("alice")

	_, err := uc.CreateDM(context.Background(), CreateDMInput{
		WorkspaceID:  "general",
		UserID:       "alice",
		TargetUserID: "outsider",
	})

	if !errors.Is(err, ErrNotWorkspaceMember) {
		t.Fatalf("ワークスペース外のユーザーとの DM が拒否されていません: %v", err)
	}
}

func TestCreateDMRejectsNonMemberRequester(t *testing.T) {
	uc := newInteractor("bob")

	_, err := uc.CreateDM(context.Background(), CreateDMInput{
		WorkspaceID:  "general",
		UserID:       "outsider",
		TargetUserID: "bob",
	})

	if !errors.Is(err, ErrNotWorkspaceMember) {
		t.Fatalf("ワークスペース外からの DM 作成が拒否されていません: %v", err)
	}
}

func TestCreateGroupDMRejectsTooManyMembers(t *testing.T) {
	uc := newInteractor()

	_, err := uc.CreateGroupDM(context.Background(), CreateGroupDMInput{
		WorkspaceID: "general",
		CreatorID:   "alice",
		MemberIDs:   []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
	})

	if !errors.Is(err, entity.ErrGroupDMMaxMembers) {
		t.Fatalf("作成者を含めた上限超過が検出されていません: %v", err)
	}
}
