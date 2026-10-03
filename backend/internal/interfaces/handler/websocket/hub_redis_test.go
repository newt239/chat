package websocket_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	"github.com/newt239/chat/internal/domain/service"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/handler/websocket"
)

// stubAccess は denied のユーザーだけチャンネルを閲覧できないものとして扱います
type stubAccess struct {
	service.ChannelAccessService
	denied string
}

func (a stubAccess) EnsureChannelAccess(_ context.Context, channelID, userID string) (*entity.Channel, error) {
	if userID == a.denied {
		return nil, domerr.ErrUnauthorized
	}
	return &entity.Channel{ID: channelID, WorkspaceID: "ws"}, nil
}

func startReplicas(t *testing.T) (*websocket.Hub, *websocket.Hub) {
	hubs := websocket.StartTestHubs(t, stubAccess{denied: "alice"}, 2)
	return hubs[0], hubs[1]
}

var testEvent = &chatv1.ServerEvent{Event: &chatv1.ServerEvent_Ack{Ack: &chatv1.AckEvent{Event: "test"}}}

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

	h1.BroadcastToChannel("ws", "general", testEvent)
	for _, c := range []*websocket.Client{alice, bob} {
		var got chatv1.ServerEvent
		if err := protojson.Unmarshal(receive(t, c), &got); err != nil || got.GetAck().GetEvent() != "test" {
			t.Fatalf("配信内容が変わっています: %v %v", &got, err)
		}
	}
	assertNothing(t, carol)

	h2.BroadcastToUsers("ws", []string{"alice"}, testEvent)
	receive(t, alice)
	assertNothing(t, bob)
}

func TestCloseEnvelopesDisconnectClientsOnOtherReplicas(t *testing.T) {
	h1, h2 := startReplicas(t)
	alice := websocket.NewTestClient(h2, "ws", "alice", "general")
	aliceOther := websocket.NewTestClient(h2, "other", "alice")
	bob := websocket.NewTestClient(h2, "ws", "bob", "general")

	h1.CloseWorkspaceUser("ws", "alice")
	assertClosed(t, alice)
	assertOpen(t, aliceOther)

	h1.CloseSession("session-bob")
	assertClosed(t, bob)

	h1.CloseUser("alice")
	assertClosed(t, aliceOther)
}

func TestRevokeChannelStopsDelivery(t *testing.T) {
	h1, h2 := startReplicas(t)
	alice := websocket.NewTestClient(h2, "ws", "alice", "general")
	bob := websocket.NewTestClient(h2, "ws", "bob", "general")

	// 閲覧できる bob は全員を確かめ直しても外れない
	h1.RevokeChannel("ws", "general", "")
	deadline := time.Now().Add(2 * time.Second)
	for {
		h1.BroadcastToChannel("ws", "general", testEvent)
		receive(t, bob)
		select {
		case <-alice.Sent():
			if time.Now().After(deadline) {
				t.Fatal("購読を外したユーザーに配信が続いています")
			}
			continue
		case <-time.After(100 * time.Millisecond):
		}
		return
	}
}

func assertClosed(t *testing.T, c *websocket.Client) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-c.Sent():
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("接続が切られませんでした")
		}
	}
}

func assertOpen(t *testing.T, c *websocket.Client) {
	t.Helper()
	select {
	case _, ok := <-c.Sent():
		if !ok {
			t.Fatal("切られないはずの接続が切られました")
		}
	case <-time.After(100 * time.Millisecond):
	}
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

func TestViewersReachViewerAndSubscribers(t *testing.T) {
	h, _ := startReplicas(t)
	alice := websocket.NewTestClient(h, "ws", "alice", "general")
	bob := websocket.NewTestClient(h, "ws", "bob")

	// 購読していない本人にも一覧が届く
	h.SetViewingChannel(bob, "general")
	lastViewers(t, bob, []string{"bob"})
	lastViewers(t, alice, []string{"bob"})

	h.SetViewingChannel(alice, "general")
	lastViewers(t, alice, []string{"alice", "bob"})

	// 別チャンネルへ移ると元のチャンネルの購読者に減った一覧が届く
	h.SetViewingChannel(bob, "random")
	lastViewers(t, alice, []string{"alice"})
}

func TestClearViewingChannelOnDisconnect(t *testing.T) {
	h, _ := startReplicas(t)
	alice := websocket.NewTestClient(h, "ws", "alice", "general")
	bob := websocket.NewTestClient(h, "ws", "bob")
	h.SetViewingChannel(alice, "general")
	h.SetViewingChannel(bob, "general")
	lastViewers(t, alice, []string{"alice", "bob"})

	bob.Disconnect()
	lastViewers(t, alice, []string{"alice"})
}
