package channelmember

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const channelID = "ch1"

type fakeMemberRepo struct {
	domainrepository.ChannelMemberRepository
	members map[string]entity.ChannelRole
}

func (r *fakeMemberRepo) AddMember(_ context.Context, m *entity.ChannelMember) error {
	if _, ok := r.members[m.UserID]; ok {
		return domerr.ErrAlreadyMember
	}
	r.members[m.UserID] = m.Role
	return nil
}

func (r *fakeMemberRepo) FindMember(_ context.Context, _, userID string) (*entity.ChannelMember, error) {
	role, ok := r.members[userID]
	if !ok {
		return nil, nil
	}
	return &entity.ChannelMember{ChannelID: channelID, UserID: userID, Role: role}, nil
}

func (r *fakeMemberRepo) CountAdmins(context.Context, string) (int, error) {
	count := 0
	for _, role := range r.members {
		if role == entity.ChannelRoleAdmin {
			count++
		}
	}
	return count, nil
}

func (r *fakeMemberRepo) RemoveMember(_ context.Context, _, userID string) error {
	delete(r.members, userID)
	return nil
}

func (r *fakeMemberRepo) UpdateMemberRole(_ context.Context, _, userID string, role entity.ChannelRole) error {
	r.members[userID] = role
	return nil
}

type stubChannelRepo struct {
	domainrepository.ChannelRepository
}

func (stubChannelRepo) FindByID(_ context.Context, id string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "ws", Type: entity.ChannelTypePrivate, CreatedBy: "owner"}, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	if userID == "outsider" {
		return nil, nil
	}
	return &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID, Role: entity.WorkspaceRoleMember}, nil
}

type stubAccess struct {
	service.ChannelAccessService
}

func (stubAccess) EnsureChannelAccess(ctx context.Context, channelID, _ string) (*entity.Channel, error) {
	return stubChannelRepo{}.FindByID(ctx, channelID)
}

type stubTx struct{}

func (stubTx) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type stubSystemMessageRepo struct {
	domainrepository.SystemMessageRepository
}

func (stubSystemMessageRepo) Create(context.Context, *entity.SystemMessage) error { return nil }

type nopNotifier struct{ messageuc.Notifier }

func (nopNotifier) NotifySystemMessageCreated(string, string, *entity.SystemMessage) {}

type stubRevoker struct{ revoked []string }

func (r *stubRevoker) RevokeChannel(_, _, userID string) { r.revoked = append(r.revoked, userID) }

func newInteractor(members map[string]entity.ChannelRole) (*Interactor, *fakeMemberRepo, *stubRevoker) {
	repo := &fakeMemberRepo{members: members}
	revoker := &stubRevoker{}
	uc := New(stubChannelRepo{}, repo, stubWorkspaceRepo{}, nil, messageuc.NewSystemMessages(stubSystemMessageRepo{}, nopNotifier{}), stubAccess{}, stubTx{}, revoker)
	return uc, repo, revoker
}

func TestLastAdminCannotLeaveOrBeDemoted(t *testing.T) {
	uc, repo, revoker := newInteractor(map[string]entity.ChannelRole{"owner": entity.ChannelRoleAdmin, "bob": entity.ChannelRoleMember})
	ctx := context.Background()

	if err := uc.LeaveChannel(ctx, channelID, "owner"); !errors.Is(err, ErrLastAdminRemoval) {
		t.Errorf("最後の管理者が退出できました: %v", err)
	}
	if err := uc.UpdateMemberRole(ctx, MemberInput{ChannelID: channelID, OperatorID: "owner", TargetUserID: "owner", Role: "member"}); !errors.Is(err, ErrLastAdminRemoval) {
		t.Errorf("最後の管理者を降格できました: %v", err)
	}
	if err := uc.UpdateMemberRole(ctx, MemberInput{ChannelID: channelID, OperatorID: "owner", TargetUserID: "bob", Role: "admin"}); err != nil {
		t.Fatalf("管理者を増やせません: %v", err)
	}
	if err := uc.LeaveChannel(ctx, channelID, "owner"); err != nil {
		t.Fatalf("管理者が他にいれば退出できるはず: %v", err)
	}
	if _, ok := repo.members["owner"]; ok || !slices.Equal(revoker.revoked, []string{"owner"}) {
		t.Errorf("退出したユーザーの配信を止めていません: %v", revoker.revoked)
	}
}

func TestInviteMember(t *testing.T) {
	uc, _, _ := newInteractor(map[string]entity.ChannelRole{"owner": entity.ChannelRoleAdmin, "bob": entity.ChannelRoleMember})
	ctx := context.Background()
	invite := func(operator, target string) error {
		return uc.InviteMember(ctx, MemberInput{ChannelID: channelID, OperatorID: operator, TargetUserID: target, Role: "member"})
	}

	if err := invite("owner", "bob"); !errors.Is(err, domerr.ErrAlreadyMember) {
		t.Errorf("参加済みのユーザーは ErrAlreadyMember のはず: %v", err)
	}
	if err := invite("owner", "outsider"); !errors.Is(err, domerr.ErrUserNotFound) {
		t.Errorf("ワークスペースのメンバー以外は招待できないはず: %v", err)
	}
	if err := invite("bob", "carol"); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Errorf("作成者でも管理者でもないユーザーは招待できないはず: %v", err)
	}
	if err := invite("owner", "carol"); err != nil {
		t.Errorf("作成者は招待できるはず: %v", err)
	}
}
