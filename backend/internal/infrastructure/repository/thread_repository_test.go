package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/newt239/chat/ent/threadreadstate"
)

func TestFollowThreadIsIdempotent(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewThreadRepository(client)
	alice, thread, other := f.alice.ID.String(), f.messages["mention"].ID.String(), f.messages["link"].ID.String()

	for range 2 {
		if err := repo.SetFollowing(ctx, alice, thread, true); err != nil {
			t.Fatalf("フォローを繰り返すと失敗しました: %v", err)
		}
	}
	followed, err := repo.FindFollowedThreadIDs(ctx, alice, []string{thread, other})
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if want := map[string]bool{thread: true}; !reflect.DeepEqual(followed, want) {
		t.Errorf("フォロー中のスレッドが期待と異なります: %v", followed)
	}

	for range 2 {
		if err := repo.SetFollowing(ctx, alice, thread, false); err != nil {
			t.Fatalf("フォロー解除を繰り返すと失敗しました: %v", err)
		}
	}
	if followed, _ := repo.FindFollowedThreadIDs(ctx, alice, []string{thread}); len(followed) != 0 {
		t.Errorf("フォローが解除されていません: %v", followed)
	}
}

func TestUpsertThreadReadState(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	ctx := context.Background()
	repo := NewThreadRepository(client)
	alice, thread := f.alice.ID.String(), f.messages["mention"].ID.String()

	later := time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)
	for _, at := range []time.Time{later.Add(-time.Hour), later} {
		if err := repo.UpsertReadState(ctx, alice, thread, at); err != nil {
			t.Fatalf("既読位置を保存できません: %v", err)
		}
	}
	states := client.ThreadReadState.Query().Where(threadreadstate.UserID(f.alice.ID)).AllX(ctx)
	if len(states) != 1 || !states[0].LastReadAt.Equal(later) {
		t.Errorf("既読位置が 1 件に更新されていません: %+v", states)
	}
}
