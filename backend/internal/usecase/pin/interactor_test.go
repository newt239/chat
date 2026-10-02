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
	"github.com/newt239/chat/internal/usecase/systemmessage"
)

type stubMessageRepo struct {
	domainrepository.MessageRepository
}

func (stubMessageRepo) FindByID(_ context.Context, id string) (*entity.Message, error) {
	return &entity.Message{ID: id, ChannelID: "ch1"}, nil
}

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

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "ws"}, nil
}

type stubPermission struct{ service.PermissionService }

func (stubPermission) Ensure(context.Context, string, string, entity.Permission) (*entity.WorkspaceMember, error) {
	return &entity.WorkspaceMember{}, nil
}

type stubSystemMessages struct{}

func (stubSystemMessages) Create(context.Context, systemmessage.CreateInput) (*entity.SystemMessage, error) {
	return &entity.SystemMessage{}, nil
}

type stubIndexer struct{}

func (stubIndexer) Sync(context.Context, ...string) {}

type recordingNotifier struct{ memberIDs [][]string }

func (n *recordingNotifier) NotifyPinCreated(_, _ string, memberIDs []string, _ PinNotification) {
	n.memberIDs = append(n.memberIDs, memberIDs)
}

func (n *recordingNotifier) NotifyPinDeleted(_, _ string, memberIDs []string, _ PinNotification) {
	n.memberIDs = append(n.memberIDs, memberIDs)
}

type nopLogger struct{ service.Logger }

func (nopLogger) Warn(string, ...service.LogField) {}

func TestPinMessageNotifiesAllMembers(t *testing.T) {
	notifier := &recordingNotifier{}
	uc := NewPinInteractor(&stubPinRepo{pinned: map[string]bool{}}, stubMessageRepo{}, stubMemberRepo{}, stubUserRepo{}, notifier, nil, stubAccess{}, stubSystemMessages{}, stubPermission{}, stubIndexer{}, nopLogger{})
	input := PinMessageInput{ChannelID: "ch1", MessageID: "m1", UserID: "alice"}

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
