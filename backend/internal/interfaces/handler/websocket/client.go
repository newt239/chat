package websocket

import (
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

const (
	// 書き込み待機時間
	writeWait = 10 * time.Second

	// 次のpingを待機する時間
	pongWait = 60 * time.Second

	// pingを送信する間隔（pongWaitより短くする必要がある）
	pingPeriod = (pongWait * 9) / 10

	// メッセージの最大サイズ
	maxMessageSize = 64 * 1024
)

// Client はWebSocket接続を表します
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	// 閲覧者の共有で接続を区別するための ID
	id string
	// 送信用のバッファードチャンネル。ハブから外すときに閉じる
	send chan []byte

	userID      string
	sessionID   string
	workspaceID string

	// 以下は Hub.mu で保護する
	subscribedChannels map[string]struct{}
	viewingChannel     string
	// 切断時に送る close フレーム。send を閉じる前に設定する
	closeMessage []byte
}

func newClient(hub *Hub, conn *websocket.Conn, id, userID, sessionID, workspaceID string) *Client {
	return &Client{
		hub:                hub,
		conn:               conn,
		id:                 id,
		send:               make(chan []byte, 256),
		userID:             userID,
		sessionID:          sessionID,
		workspaceID:        workspaceID,
		subscribedChannels: make(map[string]struct{}),
	}
}

// start はハブに登録して送受信を始めます
func (c *Client) start() {
	c.hub.register(c)
	c.hub.pumps.Go(c.writePump)
	go c.readPump()
}

// readPump はWebSocketからのメッセージを読み取ります。接続の close は writePump が行う
func (c *Client) readPump() {
	defer c.hub.unregister(c, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("WebSocket が予期せず切断されました", "user", c.userID, "error", err)
			}
			return
		}
		c.handleMessage(message)
	}
}

// writePump はWebSocketにメッセージを書き込み、送信チャンネルが閉じられたら close フレームを送って接続を閉じます
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, c.closeMessage)
				return
			}
			// クライアントは1フレーム1イベントとして解釈するため結合しない
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage はクライアントからのメッセージを処理します。不正なイベントは無視する
func (c *Client) handleMessage(data []byte) {
	var event chatv1.ClientEvent
	if err := protojson.Unmarshal(data, &event); err != nil {
		return
	}

	switch e := event.Event.(type) {
	case *chatv1.ClientEvent_JoinChannel:
		if id := e.JoinChannel.GetChannelId(); id != "" && c.hub.canAccess(c, id) {
			c.hub.subscribe(c, id)
		}
	case *chatv1.ClientEvent_LeaveChannel:
		c.hub.unsubscribe(c, e.LeaveChannel.GetChannelId())
	case *chatv1.ClientEvent_Typing:
		c.notifyTyping(e.Typing.GetChannelId(), true)
	case *chatv1.ClientEvent_StopTyping:
		c.notifyTyping(e.StopTyping.GetChannelId(), false)
	case *chatv1.ClientEvent_ViewChannel:
		// 空文字は閲覧をやめたことを表す
		if id := e.ViewChannel.GetChannelId(); id == "" || c.hub.canAccess(c, id) {
			c.hub.SetViewingChannel(c, id)
		}
	}
}

// notifyTyping は購読中のチャンネルでだけ、入力中状態の開始・停止を他の購読者に通知します
func (c *Client) notifyTyping(channelID string, typing bool) {
	if !c.hub.isSubscribed(c, channelID) {
		return
	}
	payload := &chatv1.TypingEvent{ChannelId: channelID, UserId: c.userID}
	event := &chatv1.ServerEvent{Event: &chatv1.ServerEvent_StopTyping{StopTyping: payload}}
	if typing {
		event.Event = &chatv1.ServerEvent_Typing{Typing: payload}
	}
	c.hub.broadcast(envelope{Target: targetChannel, WorkspaceID: c.workspaceID, ChannelID: channelID, ExcludeUserID: c.userID}, event)
}
