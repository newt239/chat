package redis

import (
	"slices"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/internal/domain/service"
)

func newTestClient(t *testing.T) *goredis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func TestPresenceStoreExpiresEntriesNotRefreshed(t *testing.T) {
	ctx := t.Context()
	store := NewPresenceStore(newTestClient(t))
	now := time.Unix(1000, 0)
	store.now = func() time.Time { return now }

	alicePC := service.PresenceEntry{WorkspaceID: "ws", ChannelID: "general", ConnID: "1", UserID: "alice"}
	alicePhone := service.PresenceEntry{WorkspaceID: "ws", ChannelID: "general", ConnID: "2", UserID: "alice"}
	bob := service.PresenceEntry{WorkspaceID: "ws", ChannelID: "general", ConnID: "3", UserID: "bob"}
	for _, e := range []service.PresenceEntry{alicePC, alicePhone, bob} {
		if err := store.Add(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	viewers, err := store.Viewers(ctx, "ws", "general")
	if err != nil || !slices.Equal(viewers, []string{"alice", "bob"}) {
		t.Fatalf("同じユーザーの複数接続は 1 人として数えるはず: %v %v", viewers, err)
	}

	// bob の接続を持つレプリカが落ちて延長されなくなった状態
	now = now.Add(presenceTTL - time.Second)
	if err := store.Refresh(ctx, []service.PresenceEntry{alicePC}); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Second)
	if viewers, _ := store.Viewers(ctx, "ws", "general"); !slices.Equal(viewers, []string{"alice"}) {
		t.Fatalf("延長されなかった閲覧が残っています: %v", viewers)
	}

	if err := store.Remove(ctx, alicePC); err != nil {
		t.Fatal(err)
	}
	if viewers, _ := store.Viewers(ctx, "ws", "general"); len(viewers) != 0 {
		t.Fatalf("閲覧をやめた接続が残っています: %v", viewers)
	}
}
