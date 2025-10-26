package websocket

import (
	"testing"
	"time"
)

func TestHub_RegisterAndUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// モッククライアントを作成
	client := &Client{
		hub:         hub,
		conn:        nil, // テストではnilで許容
		send:        make(chan []byte, 256),
		userID:      "user1",
		workspaceID: "workspace1",
	}

	// クライアントを登録
	hub.register <- client

	// 少し待機して登録を確認
	time.Sleep(10 * time.Millisecond)

	// 接続ユーザーを確認
	users := hub.GetConnectedUsers("workspace1")
	if len(users) != 1 {
		t.Errorf("Expected 1 connected user, got %d", len(users))
	}
	if users[0] != "user1" {
		t.Errorf("Expected user1, got %s", users[0])
	}

	// クライアントを登録解除
	hub.unregister <- client

	// 少し待機して登録解除を確認
	time.Sleep(10 * time.Millisecond)

	// 接続ユーザーがいないことを確認
	users = hub.GetConnectedUsers("workspace1")
	if len(users) != 0 {
		t.Errorf("Expected 0 connected users, got %d", len(users))
	}
}

func TestHub_MultipleConnections(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 同一ユーザーの複数接続を作成
	client1 := &Client{
		hub:         hub,
		conn:        nil,
		send:        make(chan []byte, 256),
		userID:      "user1",
		workspaceID: "workspace1",
	}

	client2 := &Client{
		hub:         hub,
		conn:        nil,
		send:        make(chan []byte, 256),
		userID:      "user1",
		workspaceID: "workspace1",
	}

	// 両方のクライアントを登録
	hub.register <- client1
	hub.register <- client2

	time.Sleep(10 * time.Millisecond)

	// ユーザー数は1のまま（同一ユーザー）
	users := hub.GetConnectedUsers("workspace1")
	if len(users) != 1 {
		t.Errorf("Expected 1 unique user, got %d", len(users))
	}

	// 1つ目の接続を解除
	hub.unregister <- client1
	time.Sleep(10 * time.Millisecond)

	// まだユーザーは接続中（client2が残っている）
	users = hub.GetConnectedUsers("workspace1")
	if len(users) != 1 {
		t.Errorf("Expected 1 user still connected, got %d", len(users))
	}

	// 2つ目の接続も解除
	hub.unregister <- client2
	time.Sleep(10 * time.Millisecond)

	// 全て解除された
	users = hub.GetConnectedUsers("workspace1")
	if len(users) != 0 {
		t.Errorf("Expected 0 users, got %d", len(users))
	}
}

func TestHub_BroadcastToWorkspace(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// 2つのクライアントを作成
	client1 := &Client{
		hub:         hub,
		conn:        nil,
		send:        make(chan []byte, 256),
		userID:      "user1",
		workspaceID: "workspace1",
	}

	client2 := &Client{
		hub:         hub,
		conn:        nil,
		send:        make(chan []byte, 256),
		userID:      "user2",
		workspaceID: "workspace1",
	}

	hub.register <- client1
	hub.register <- client2

	time.Sleep(10 * time.Millisecond)

	// ワークスペース全体にブロードキャスト
	message := []byte(`{"type":"test","payload":{}}`)
	hub.BroadcastToWorkspace("workspace1", message)

	time.Sleep(10 * time.Millisecond)

	// 両方のクライアントがメッセージを受信したことを確認
	select {
	case msg := <-client1.send:
		if string(msg) != string(message) {
			t.Errorf("Client1 received wrong message: %s", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive message")
	}

	select {
	case msg := <-client2.send:
		if string(msg) != string(message) {
			t.Errorf("Client2 received wrong message: %s", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client2 did not receive message")
	}
}

func TestHub_BroadcastToUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client1 := &Client{
		hub:         hub,
		conn:        nil,
		send:        make(chan []byte, 256),
		userID:      "user1",
		workspaceID: "workspace1",
	}

	client2 := &Client{
		hub:         hub,
		conn:        nil,
		send:        make(chan []byte, 256),
		userID:      "user2",
		workspaceID: "workspace1",
	}

	hub.register <- client1
	hub.register <- client2

	time.Sleep(10 * time.Millisecond)

	// user1のみにメッセージを送信
	message := []byte(`{"type":"test","payload":{}}`)
	hub.BroadcastToUser("workspace1", "user1", message)

	time.Sleep(10 * time.Millisecond)

	// client1のみがメッセージを受信
	select {
	case msg := <-client1.send:
		if string(msg) != string(message) {
			t.Errorf("Client1 received wrong message: %s", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client1 did not receive message")
	}

	// client2はメッセージを受信しない
	select {
	case <-client2.send:
		t.Error("Client2 should not receive message")
	case <-time.After(50 * time.Millisecond):
		// 期待通り
	}
}
