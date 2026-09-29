package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestShutdownClosesConnectionsWithGoingAway(t *testing.T) {
	h := NewHub()
	go h.Run(t.Context())

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := &Client{hub: h, conn: conn, id: "1", send: make(chan []byte, 8), userID: "alice", workspaceID: "ws", subscribedChannels: map[string]bool{}}
		h.register <- c
		go c.writePump()
		go c.readPump()
	}))
	defer srv.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	for len(h.clients()) == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	done := make(chan struct{})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.Shutdown(ctx)
		close(done)
	}()

	// クライアントは close フレームに応答し、1001 を見て他のレプリカへつなぎ直す
	if _, _, err := conn.ReadMessage(); !websocket.IsCloseError(err, websocket.CloseGoingAway) {
		t.Fatalf("Going Away で閉じられるはず: %v", err)
	}
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("クライアントが切断しても Shutdown が終わりません")
	}
	if n := len(h.clients()); n != 0 {
		t.Fatalf("接続が残っています: %d", n)
	}
}
