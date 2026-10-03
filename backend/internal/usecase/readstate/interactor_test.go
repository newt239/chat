package readstate

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
)

type stubAccess struct {
	domainservice.ChannelAccessService
	descendants []*entity.Channel
}

func (stubAccess) EnsureChannelAccess(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id}, nil
}

func (s stubAccess) AccessibleDescendants(_ context.Context, _ *entity.Channel, _ string) ([]*entity.Channel, error) {
	return s.descendants, nil
}

type fakeReadStateRepo struct {
	domainrepository.ReadStateRepository
	lastReadAt map[string]time.Time
}

func (r *fakeReadStateRepo) AdvanceBatch(_ context.Context, channelIDs []string, _ string, lastReadAt time.Time) error {
	for _, id := range channelIDs {
		if at, ok := r.lastReadAt[id]; !ok || at.Before(lastReadAt) {
			r.lastReadAt[id] = lastReadAt
		}
	}
	return nil
}

func (r *fakeReadStateRepo) GetUnreadCountBatch(context.Context, []string, string) (map[string]int, error) {
	return map[string]int{}, nil
}

func (r *fakeReadStateRepo) GetUnreadMentionCountBatch(context.Context, []string, string) (map[string]int, error) {
	return map[string]int{}, nil
}

type recordingNotifier struct{ notified []string }

func (n *recordingNotifier) NotifyUnreadCount(_, _, channelID string, _, _ int) {
	n.notified = append(n.notified, channelID)
}

func (r *fakeReadStateRepo) Upsert(_ context.Context, rs *entity.ChannelReadState) error {
	r.lastReadAt[rs.ChannelID] = rs.LastReadAt
	return nil
}

func TestUpdateReadStateIncludeDescendants(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	repo := &fakeReadStateRepo{lastReadAt: map[string]time.Time{"newer": later}}
	access := stubAccess{descendants: []*entity.Channel{{ID: "unread"}, {ID: "newer"}}}
	notifier := &recordingNotifier{}
	uc := New(repo, notifier, access)

	if err := uc.UpdateReadState(context.Background(), UpdateReadStateInput{ChannelID: "parent", UserID: "u", LastReadAt: now, IncludeDescendants: true}); err != nil {
		t.Fatalf("既読にできません: %v", err)
	}

	if !repo.lastReadAt["parent"].Equal(now) || !repo.lastReadAt["unread"].Equal(now) {
		t.Fatal("親と子孫が既読になっていません")
	}
	if !repo.lastReadAt["newer"].Equal(later) {
		t.Fatal("子孫の既読位置が巻き戻っています")
	}
	if len(notifier.notified) != 3 {
		t.Errorf("既読にしたチャンネルの未読数を配信していません: %v", notifier.notified)
	}
}

func TestUpdateReadStateWithoutDescendants(t *testing.T) {
	repo := &fakeReadStateRepo{lastReadAt: map[string]time.Time{}}
	access := stubAccess{descendants: []*entity.Channel{{ID: "child"}}}
	uc := New(repo, &recordingNotifier{}, access)

	if err := uc.UpdateReadState(context.Background(), UpdateReadStateInput{ChannelID: "parent", UserID: "u", LastReadAt: time.Now()}); err != nil {
		t.Fatal(err)
	}

	if _, ok := repo.lastReadAt["child"]; ok {
		t.Fatal("指定していないのに子孫が既読になっています")
	}
}
