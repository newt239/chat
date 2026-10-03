package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

// dialTestClient は実際の WebSocket 接続を 1 本張り、ハブに登録します
func dialTestClient(t *testing.T, h *Hub, sessionID string) *websocket.Conn {
	t.Helper()
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		newClient(h, conn, sessionID, "alice", sessionID, "ws").start()
	}))
	t.Cleanup(srv.Close)

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	for len(h.clients()) == 0 {
		time.Sleep(10 * time.Millisecond)
	}
	return conn
}

func expectSent(t *testing.T, c *Client) {
	t.Helper()
	select {
	case <-c.send:
	case <-time.After(2 * time.Second):
		t.Fatal("イベントが届きませんでした")
	}
}

func expectNothing(t *testing.T, c *Client) {
	t.Helper()
	select {
	case data := <-c.send:
		t.Fatalf("届かないはずのイベントが届きました: %s", data)
	case <-time.After(100 * time.Millisecond):
	}
}

var testEvent = &chatv1.ServerEvent{Event: &chatv1.ServerEvent_Ack{Ack: &chatv1.AckEvent{Event: "test"}}}

func TestShutdownClosesConnectionsWithGoingAway(t *testing.T) {
	h := StartTestHubs(t, nil, 1)[0]
	conn := dialTestClient(t, h, "s1")

	done := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.Shutdown(ctx)
		close(done)
	}()

	// クライアントは 1001 を見て他のレプリカへつなぎ直す
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseGoingAway) {
		t.Fatalf("Going Away で閉じられるはず: %v", err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Shutdown が終わりません")
	}
	if n := len(h.clients()); n != 0 {
		t.Fatalf("接続が残っています: %d", n)
	}
}

func TestCloseSessionSendsUnauthenticatedCode(t *testing.T) {
	h := StartTestHubs(t, nil, 1)[0]
	conn := dialTestClient(t, h, "s1")

	h.CloseSession("s1")
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, CloseUnauthenticated) {
		t.Fatalf("4401 で閉じられるはず: %v", err)
	}
}

func TestSubscriptionsArePerConnection(t *testing.T) {
	h := StartTestHubs(t, nil, 1)[0]
	tab1 := NewTestClient(h, "ws", "alice", "general")
	tab2 := NewTestClient(h, "ws", "alice")

	h.BroadcastToChannel("ws", "general", testEvent)
	expectSent(t, tab1)
	expectNothing(t, tab2)

	h.unsubscribe(tab1, "general")
	h.BroadcastToChannel("ws", "general", testEvent)
	expectNothing(t, tab1)
}

func TestTypingRequiresSubscription(t *testing.T) {
	h := StartTestHubs(t, nil, 1)[0]
	alice := NewTestClient(h, "ws", "alice")
	bob := NewTestClient(h, "ws", "bob", "general")

	alice.notifyTyping("general", true)
	expectNothing(t, bob)
	h.subscribe(alice, "general")
	<-alice.send
	alice.notifyTyping("general", true)
	expectSent(t, bob)
	// 入力中の通知は本人には届かない
	expectNothing(t, alice)
}
