package pin

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type stubPinRepo struct {
	domainrepository.PinRepository
	pinned map[string]bool
}

func (r *stubPinRepo) Create(_ context.Context, p *entity.MessagePin) error {
	if r.pinned[p.MessageID] {
		return domerr.ErrPinExists
	}
	r.pinned[p.MessageID] = true
	p.PinnedAt = time.Now()
	return nil
}

type stubMemberRepo struct {
	domainrepository.ChannelMemberRepository
}

func (stubMemberRepo) FindMembers(context.Context, string) ([]*entity.ChannelMember, error) {
	return []*entity.ChannelMember{{UserID: "alice"}, {UserID: "bob"}}, nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByID(_ context.Context, id string) (*entity.User, error) {
	return &entity.User{ID: id, DisplayName: id}, nil
}

type stubAccess struct{ service.ChannelAccessService }

func (stubAccess) EnsureMessageAccess(_ context.Context, id string, _ string) (*entity.Message, *entity.Channel, error) {
	return &entity.Message{ID: id, ChannelID: "ch1"}, &entity.Channel{ID: "ch1", WorkspaceID: "ws"}, nil
}

type stubPermission struct{ service.PermissionService }

func (stubPermission) Ensure(context.Context, string, string, entity.Permission) (*entity.WorkspaceMember, error) {
	return &entity.WorkspaceMember{}, nil
}

type stubSystemMessageRepo struct {
	domainrepository.SystemMessageRepository
}

func (stubSystemMessageRepo) Create(context.Context, *entity.SystemMessage) error { return nil }

type nopNotifier struct{ messageuc.Notifier }

func (nopNotifier) NotifySystemMessageCreated(string, string, *entity.SystemMessage) {}

type stubIndexer struct{}

func (stubIndexer) Sync(context.Context, ...string) {}

type recordingNotifier struct{ memberIDs [][]string }

func (n *recordingNotifier) NotifyPinCreated(_, _ string, memberIDs []string, _ PinNotification) {
	n.memberIDs = append(n.memberIDs, memberIDs)
}

func (n *recordingNotifier) NotifyPinDeleted(_, _ string, memberIDs []string, _ PinNotification) {
	n.memberIDs = append(n.memberIDs, memberIDs)
}

func TestPinMessageNotifiesAllMembers(t *testing.T) {
	notifier := &recordingNotifier{}
	uc := New(&stubPinRepo{pinned: map[string]bool{}}, stubMemberRepo{}, stubUserRepo{}, notifier, nil, stubAccess{}, messageuc.NewSystemMessages(stubSystemMessageRepo{}, nopNotifier{}), stubPermission{}, stubIndexer{})
	input := PinInput{ChannelID: "ch1", MessageID: "m1", UserID: "alice"}

	if err := uc.PinMessage(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if err := uc.PinMessage(context.Background(), input); !errors.Is(err, domerr.ErrPinExists) {
		t.Errorf("二重のピン留めは ErrPinExists のはず: %v", err)
	}
	if len(notifier.memberIDs) != 1 || !slices.Equal(notifier.memberIDs[0], []string{"alice", "bob"}) {
		t.Errorf("チャンネルの参加者全員に知らせていません: %v", notifier.memberIDs)
	}
}
