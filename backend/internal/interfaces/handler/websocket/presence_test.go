package websocket

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

func newTestClient(h *Hub, userID string) *Client {
	c := &Client{hub: h, id: userID + "-conn", send: make(chan []byte, 8), userID: userID, workspaceID: "ws"}
	if h.workspaces["ws"] == nil {
		h.workspaces["ws"] = map[string][]*Client{}
	}
	h.workspaces["ws"][userID] = append(h.workspaces["ws"][userID], c)
	return c
}

// lastViewers は受信したイベントのうち最後の閲覧者一覧を返します
func lastViewers(t *testing.T, c *Client) *chatv1.ChannelViewersEvent {
	t.Helper()
	var last *chatv1.ChannelViewersEvent
	for {
		select {
		case data := <-c.send:
			var event chatv1.ServerEvent
			if err := protojson.Unmarshal(data, &event); err != nil {
				t.Fatal(err)
			}
			last = event.GetChannelViewers()
		default:
			return last
		}
	}
}

func TestSetViewingChannelBroadcastsViewers(t *testing.T) {
	h := NewHub()
	alice := newTestClient(h, "alice")
	bob := newTestClient(h, "bob")
	h.channelSubscribers["ws"] = map[string]map[string]bool{"general": {"alice": true}}

	// 購読していない本人にも一覧が届く
	h.SetViewingChannel(bob, "general")
	if got := lastViewers(t, bob); got == nil || !slices.Equal(got.UserIds, []string{"bob"}) {
		t.Fatalf("閲覧を始めた本人に一覧が届いていません: %v", got)
	}
	if got := lastViewers(t, alice); got == nil || !slices.Equal(got.UserIds, []string{"bob"}) {
		t.Fatalf("購読者に一覧が届いていません: %v", got)
	}

	h.SetViewingChannel(alice, "general")
	if got := lastViewers(t, alice); !slices.Equal(got.UserIds, []string{"alice", "bob"}) {
		t.Fatalf("閲覧者が増えていません: %v", got.UserIds)
	}

	// 別チャンネルへ移ると元のチャンネルの購読者に減った一覧が届く
	h.SetViewingChannel(bob, "random")
	if got := lastViewers(t, alice); got.ChannelId != "general" || !slices.Equal(got.UserIds, []string{"alice"}) {
		t.Fatalf("閲覧をやめたユーザーが残っています: %v", got)
	}
}

func TestClearViewingChannelOnDisconnect(t *testing.T) {
	h := NewHub()
	alice := newTestClient(h, "alice")
	bob := newTestClient(h, "bob")
	h.channelSubscribers["ws"] = map[string]map[string]bool{"general": {"alice": true}}
	h.SetViewingChannel(alice, "general")
	h.SetViewingChannel(bob, "general")
	lastViewers(t, alice)

	h.workspaces["ws"]["bob"] = nil
	bob.viewingChannel = ""
	h.leaveViewing(bob, "general")

	if got := lastViewers(t, alice); !slices.Equal(got.UserIds, []string{"alice"}) {
		t.Fatalf("切断したユーザーが閲覧者に残っています: %v", got.UserIds)
	}
}
