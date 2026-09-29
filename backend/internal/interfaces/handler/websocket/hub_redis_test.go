package websocket_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/infrastructure/redis"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
)

// startReplicas は同じ Redis を共有する 2 つのハブを、レプリカに見立てて起動します
func startReplicas(t *testing.T) (*websocket.Hub, *websocket.Hub) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	hubs := make([]*websocket.Hub, 2)
	for i := range hubs {
		hubs[i] = websocket.NewHub(websocket.WithBroker(redis.NewBroker(rdb)), websocket.WithPresenceStore(redis.NewPresenceStore(rdb)))
		go hubs[i].Run(ctx)
	}

	deadline := time.Now().Add(2 * time.Second)
	for mr.PubSubNumSub("chat:ws:events")["chat:ws:events"] < len(hubs) {
		if time.Now().After(deadline) {
			t.Fatal("ハブが Redis を購読しませんでした")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return hubs[0], hubs[1]
}

func receive(t *testing.T, c *websocket.Client) []byte {
	t.Helper()
	select {
	case data := <-c.Sent():
		return data
	case <-time.After(2 * time.Second):
		t.Fatal("イベントが届きませんでした")
		return nil
	}
}

func assertNothing(t *testing.T, c *websocket.Client) {
	t.Helper()
	select {
	case data := <-c.Sent():
		t.Fatalf("届かないはずのイベントが届きました: %s", data)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestBroadcastReachesClientsOnOtherReplicas(t *testing.T) {
	h1, h2 := startReplicas(t)
	alice := websocket.NewTestClient(h1, "ws", "alice", "general")
	bob := websocket.NewTestClient(h2, "ws", "bob", "general")
	carol := websocket.NewTestClient(h2, "ws", "carol")

	h1.BroadcastToChannelSubscribers("ws", "general", []byte(`{"n":1}`))
	for _, c := range []*websocket.Client{alice, bob} {
		if got := string(receive(t, c)); got != `{"n":1}` {
			t.Fatalf("配信内容が変わっています: %s", got)
		}
	}
	assertNothing(t, carol)

	// 入力中の通知は本人を除いて届く
	h2.BroadcastToChannel("ws", "general", []byte(`{"n":2}`), "alice")
	receive(t, bob)
	assertNothing(t, alice)

	h2.BroadcastToUser("ws", "alice", []byte(`{"n":3}`))
	receive(t, alice)
	assertNothing(t, bob)
}

func lastViewers(t *testing.T, c *websocket.Client, want []string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case data := <-c.Sent():
			var event chatv1.ServerEvent
			if err := protojson.Unmarshal(data, &event); err != nil {
				t.Fatal(err)
			}
			if slices.Equal(event.GetChannelViewers().GetUserIds(), want) {
				return
			}
		case <-deadline:
			t.Fatalf("閲覧者一覧 %v が届きませんでした", want)
		}
	}
}

func TestViewersAreAggregatedAcrossReplicas(t *testing.T) {
	h1, h2 := startReplicas(t)
	alice := websocket.NewTestClient(h1, "ws", "alice", "general")
	bob := websocket.NewTestClient(h2, "ws", "bob", "general")

	h1.SetViewingChannel(alice, "general")
	lastViewers(t, bob, []string{"alice"})

	h2.SetViewingChannel(bob, "general")
	lastViewers(t, alice, []string{"alice", "bob"})

	h2.SetViewingChannel(bob, "")
	lastViewers(t, alice, []string{"alice"})
}
