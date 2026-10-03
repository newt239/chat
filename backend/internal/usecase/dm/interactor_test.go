package dm

import (
	"context"
	"errors"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/domain/repository"
)

type stubWorkspaceRepo struct {
	repository.WorkspaceRepository
	members map[string]*entity.WorkspaceMember
}

func (r *stubWorkspaceRepo) FindActiveMemberIDs(_ context.Context, _ string, userIDs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range userIDs {
		if r.members[id] != nil {
			result[id] = true
		}
	}
	return result, nil
}

func newInteractor(memberIDs ...string) *Interactor {
	members := make(map[string]*entity.WorkspaceMember, len(memberIDs))
	for _, id := range memberIDs {
		members[id] = &entity.WorkspaceMember{Role: entity.WorkspaceRoleMember}
	}
	return NewInteractor(nil, nil, nil, nil, nil, nil, &stubWorkspaceRepo{members: members})
}

func TestCreateDMRejectsNonWorkspaceMember(t *testing.T) {
	uc := newInteractor("alice")

	_, err := uc.CreateDM(context.Background(), CreateDMInput{
		WorkspaceID:  "general",
		UserID:       "alice",
		TargetUserID: "outsider",
	})

	if !errors.Is(err, domerr.ErrUnauthorized) {
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

	if !errors.Is(err, domerr.ErrUnauthorized) {
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

type stubDMChannelRepo struct{ repository.ChannelRepository }

func (stubDMChannelRepo) FindUserDMs(context.Context, string, string) ([]*entity.Channel, error) {
	return []*entity.Channel{{ID: "dm1", Type: entity.ChannelTypeDM}, {ID: "g1", Type: entity.ChannelTypeGroupDM}}, nil
}

type stubMemberRepo struct {
	repository.ChannelMemberRepository
}

func (stubMemberRepo) FindMembersByChannelIDs(context.Context, []string) ([]*entity.ChannelMember, error) {
	return []*entity.ChannelMember{
		{ChannelID: "dm1", UserID: "alice"}, {ChannelID: "dm1", UserID: "bob"},
		{ChannelID: "g1", UserID: "alice"}, {ChannelID: "g1", UserID: "bob"}, {ChannelID: "g1", UserID: "carol"},
	}, nil
}

type stubUserRepo struct {
	repository.UserRepository
	calls int
}

func (r *stubUserRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.User, error) {
	r.calls++
	users := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, &entity.User{ID: id, DisplayName: id})
	}
	return users, nil
}

type stubFlags struct {
	repository.ChannelStarRepository
	repository.ChannelMuteRepository
	repository.ReadStateRepository
}

func (stubFlags) FindStarredChannelIDs(context.Context, string, []string) (map[string]bool, error) {
	return map[string]bool{"g1": true}, nil
}

func (stubFlags) FindMutedChannelIDs(context.Context, string, []string) (map[string]bool, error) {
	return map[string]bool{}, nil
}

func (stubFlags) GetUnreadCountBatch(context.Context, []string, string) (map[string]int, error) {
	return map[string]int{"dm1": 2}, nil
}

func (stubFlags) GetUnreadMentionCountBatch(context.Context, []string, string) (map[string]int, error) {
	return map[string]int{}, nil
}

func TestListDMsLoadsMembersAtOnce(t *testing.T) {
	users := &stubUserRepo{}
	uc := NewInteractor(stubDMChannelRepo{}, stubMemberRepo{}, stubFlags{}, stubFlags{}, stubFlags{}, users, nil)

	dms, err := uc.ListDMs(context.Background(), ListDMsInput{WorkspaceID: "ws", UserID: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if users.calls != 1 {
		t.Errorf("参加者をまとめて読み込んでいません: %d 回", users.calls)
	}
	if len(dms[0].Members) != 1 || dms[0].Members[0].UserID != "bob" || dms[0].UnreadCount != 2 {
		t.Errorf("DM の内容が期待と異なります: %+v", dms[0])
	}
	if len(dms[1].Members) != 2 || !dms[1].IsStarred {
		t.Errorf("グループ DM の内容が期待と異なります: %+v", dms[1])
	}
}
