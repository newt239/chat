package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/infrastructure/redis"
)

// StartTestHubs は同じ Redis を共有する n 個のハブを、レプリカに見立てて起動します
func StartTestHubs(t *testing.T, access service.ChannelAccessService, n int) []*Hub {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	hubs := make([]*Hub, n)
	for i := range hubs {
		hubs[i] = NewHub(access, redis.NewBroker(rdb), redis.NewPresenceStore(rdb))
		go hubs[i].Run(ctx)
	}

	deadline := time.Now().Add(2 * time.Second)
	for mr.PubSubNumSub("chat:ws:events")["chat:ws:events"] < n {
		if time.Now().After(deadline) {
			t.Fatal("ハブが Redis を購読しませんでした")
		}
		time.Sleep(10 * time.Millisecond)
	}
	return hubs
}

// NewTestClient は接続を持たないクライアントをハブに登録し、指定したチャンネルを購読させます
func NewTestClient(h *Hub, workspaceID, userID string, channels ...string) *Client {
	c := newClient(h, nil, userID+"@"+workspaceID, userID, "session-"+userID, workspaceID)
	c.send = make(chan []byte, 32)
	h.register(c)
	for _, ch := range channels {
		h.subscribe(c, ch)
	}
	return c
}

// Sent はクライアントに送られたデータを返します
func (c *Client) Sent() <-chan []byte {
	return c.send
}

// Disconnect は接続が切れたときと同じくハブから外します
func (c *Client) Disconnect() {
	c.hub.unregister(c, nil)
}
