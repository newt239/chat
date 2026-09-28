package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/newt239/chat/internal/domain/service"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

// Hub はWebSocket接続を管理します
type Hub struct {
	// workspaces と channelSubscribers を保護する
	mu sync.RWMutex

	// Workspace単位でクライアントを管理
	// workspaceID -> userID -> []*Client (同一ユーザーの複数接続をサポート)
	workspaces map[string]map[string][]*Client

	// チャンネルごとの購読者を管理
	// workspaceID -> channelID -> userID -> bool
	channelSubscribers map[string]map[string]map[string]bool

	// クライアントからの登録要求
	register chan *Client

	// クライアントからの登録解除要求
	unregister chan *Client

	// ブロードキャスト用のチャンネル
	broadcast chan *BroadcastMessage

	// チャンネル購読管理用チャンネル
	subscribe   chan *SubscribeRequest
	unsubscribe chan *UnsubscribeRequest
}

// SubscribeRequest はチャンネル購読リクエストを表します
type SubscribeRequest struct {
	WorkspaceID string
	ChannelID   string
	UserID      string
}

// UnsubscribeRequest はチャンネル購読解除リクエストを表します
type UnsubscribeRequest struct {
	WorkspaceID string
	ChannelID   string
	UserID      string
}

// BroadcastMessage はブロードキャストメッセージを表します
type BroadcastMessage struct {
	WorkspaceID string
	ChannelID   *string // nilの場合はWorkspace全体にブロードキャスト
	ExcludeUser *string // 特定ユーザーを除外する場合
	Data        []byte
}

// Client はWebSocket接続を表します
type Client struct {
	// WebSocketハブ
	hub *Hub

	// WebSocket接続
	conn *websocket.Conn

	// 送信用のバッファードチャンネル
	send chan []byte

	// ユーザーID
	userID string

	// ワークスペースID
	workspaceID string

	// 購読中のチャンネルID一覧
	subscribedChannels map[string]bool

	// 閲覧中のチャンネルID。Hub.mu で保護する
	viewingChannel string

	// チャンネルの閲覧権限の確認に使う
	channelAccess service.ChannelAccessService
}

// NewHub は新しいHubを作成します
func NewHub() *Hub {
	return &Hub{
		workspaces:         make(map[string]map[string][]*Client),
		channelSubscribers: make(map[string]map[string]map[string]bool),
		register:           make(chan *Client),
		unregister:         make(chan *Client, 256),
		broadcast:          make(chan *BroadcastMessage, 256),
		subscribe:          make(chan *SubscribeRequest),
		unsubscribe:        make(chan *UnsubscribeRequest),
	}
}

// Run はハブを開始します
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Workspaceが存在しない場合は作成
			if h.workspaces[client.workspaceID] == nil {
				h.workspaces[client.workspaceID] = make(map[string][]*Client)
			}
			// ユーザーのクライアントリストに追加（複数接続をサポート）
			h.workspaces[client.workspaceID][client.userID] = append(
				h.workspaces[client.workspaceID][client.userID],
				client,
			)
			log.Printf("[WebSocket] クライアント登録: user=%s workspace=%s 接続数=%d",
				client.userID, client.workspaceID, len(h.workspaces[client.workspaceID][client.userID]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if workspace, ok := h.workspaces[client.workspaceID]; ok {
				if clients, ok := workspace[client.userID]; ok {
					// クライアントリストから削除
					for i, c := range clients {
						if c == client {
							workspace[client.userID] = append(clients[:i], clients[i+1:]...)
							close(client.send)
							h.clearViewingChannel(client)
							break
						}
					}
					// クライアントがいなくなったらユーザーを削除
					if len(workspace[client.userID]) == 0 {
						delete(workspace, client.userID)
						// このユーザーが購読していた全チャンネルから削除
						h.removeUserFromAllChannels(client.workspaceID, client.userID)
					}
					// Workspaceにユーザーがいなくなったら削除
					if len(workspace) == 0 {
						delete(h.workspaces, client.workspaceID)
					}
					log.Printf("[WebSocket] クライアント登録解除: user=%s workspace=%s 残接続数=%d",
						client.userID, client.workspaceID, len(workspace[client.userID]))
				}
			}
			h.mu.Unlock()

		case req := <-h.subscribe:
			h.mu.Lock()
			// チャンネル購読者リストに追加
			if h.channelSubscribers[req.WorkspaceID] == nil {
				h.channelSubscribers[req.WorkspaceID] = make(map[string]map[string]bool)
			}
			if h.channelSubscribers[req.WorkspaceID][req.ChannelID] == nil {
				h.channelSubscribers[req.WorkspaceID][req.ChannelID] = make(map[string]bool)
			}
			h.channelSubscribers[req.WorkspaceID][req.ChannelID][req.UserID] = true
			log.Printf("[WebSocket] チャンネル購読者登録: user=%s workspace=%s channel=%s",
				req.UserID, req.WorkspaceID, req.ChannelID)
			h.mu.Unlock()

		case req := <-h.unsubscribe:
			h.mu.Lock()
			// チャンネル購読者リストから削除
			if wsChannels, ok := h.channelSubscribers[req.WorkspaceID]; ok {
				if subscribers, ok := wsChannels[req.ChannelID]; ok {
					delete(subscribers, req.UserID)
					if len(subscribers) == 0 {
						delete(wsChannels, req.ChannelID)
					}
					log.Printf("[WebSocket] チャンネル購読者解除: user=%s workspace=%s channel=%s",
						req.UserID, req.WorkspaceID, req.ChannelID)
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if workspace, ok := h.workspaces[msg.WorkspaceID]; ok {
				for userID, clients := range workspace {
					// ExcludeUserが設定されている場合はスキップ
					if msg.ExcludeUser != nil && userID == *msg.ExcludeUser {
						continue
					}

					// ChannelIDが指定されている場合は購読チェック
					if msg.ChannelID != nil {
						// そのチャンネルを購読しているかチェック
						if !h.isUserSubscribedToChannel(msg.WorkspaceID, *msg.ChannelID, userID) {
							continue
						}
					}

					for _, client := range clients {
						h.trySend(client, msg.Data)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// trySend は送信バッファが詰まっている接続を切断対象にします
// close は unregister 経路に一本化し、二重 close による panic を防ぎます
func (h *Hub) trySend(client *Client, data []byte) {
	select {
	case client.send <- data:
	default:
		select {
		case h.unregister <- client:
		default:
		}
	}
}

// removeUserFromAllChannels はユーザーが購読している全チャンネルから削除します
func (h *Hub) removeUserFromAllChannels(workspaceID, userID string) {
	if wsChannels, ok := h.channelSubscribers[workspaceID]; ok {
		for channelID, subscribers := range wsChannels {
			delete(subscribers, userID)
			if len(subscribers) == 0 {
				delete(wsChannels, channelID)
			}
		}
	}
}

// isUserSubscribedToChannel はユーザーが指定されたチャンネルを購読しているかチェックします
func (h *Hub) isUserSubscribedToChannel(workspaceID, channelID, userID string) bool {
	if wsChannels, ok := h.channelSubscribers[workspaceID]; ok {
		if subscribers, ok := wsChannels[channelID]; ok {
			return subscribers[userID]
		}
	}
	return false
}

// BroadcastToWorkspace はWorkspace内の全クライアントにメッセージを送信します
func (h *Hub) BroadcastToWorkspace(workspaceID string, message []byte) {
	h.broadcast <- &BroadcastMessage{
		WorkspaceID: workspaceID,
		Data:        message,
	}
	log.Printf("[WebSocket] Workspaceブロードキャスト: workspace=%s サイズ=%d bytes", workspaceID, len(message))
}

// BroadcastToChannel はChannel内の全クライアントにメッセージを送信します
// excludeUserID が空でない場合はそのユーザーを配信対象から除外します
func (h *Hub) BroadcastToChannel(workspaceID string, channelID string, message []byte, excludeUserID string) {
	var exclude *string
	if excludeUserID != "" {
		exclude = &excludeUserID
	}

	h.broadcast <- &BroadcastMessage{
		WorkspaceID: workspaceID,
		ChannelID:   &channelID,
		ExcludeUser: exclude,
		Data:        message,
	}
	log.Printf("[WebSocket] Channelブロードキャスト: workspace=%s channel=%s サイズ=%d bytes",
		workspaceID, channelID, len(message))
}

// BroadcastToUser は特定のユーザーにメッセージを送信します
func (h *Hub) BroadcastToUser(workspaceID string, userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if workspace, ok := h.workspaces[workspaceID]; ok {
		if clients, ok := workspace[userID]; ok {
			for _, client := range clients {
				h.trySend(client, message)
			}
			log.Printf("[WebSocket] ユーザー宛送信: workspace=%s user=%s 接続数=%d サイズ=%d bytes",
				workspaceID, userID, len(clients), len(message))
		}
	}
}

// BroadcastToChannelSubscribers はチャンネルを購読している全ユーザーにメッセージを送信します
// メッセージイベント(新着/編集/削除)の配信に使用します
func (h *Hub) BroadcastToChannelSubscribers(workspaceID string, channelID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// チャンネルの購読者を取得
	var subscribers []string
	if wsChannels, ok := h.channelSubscribers[workspaceID]; ok {
		if subs, ok := wsChannels[channelID]; ok {
			subscribers = make([]string, 0, len(subs))
			for userID := range subs {
				subscribers = append(subscribers, userID)
			}
		}
	}

	if len(subscribers) == 0 {
		log.Printf("[WebSocket] 購読者なし: workspace=%s channel=%s", workspaceID, channelID)
		return
	}

	// 購読者全員にメッセージを送信
	if workspace, ok := h.workspaces[workspaceID]; ok {
		sentCount := 0
		for _, userID := range subscribers {
			if clients, ok := workspace[userID]; ok {
				for _, client := range clients {
					h.trySend(client, message)
					sentCount++
				}
			}
		}
		log.Printf("[WebSocket] 購読者向けブロードキャスト: workspace=%s channel=%s 購読者数=%d 送信数=%d サイズ=%d bytes",
			workspaceID, channelID, len(subscribers), sentCount, len(message))
	}
}

// GetConnectedUsers は指定されたWorkspace内の接続中のユーザーIDリストを返します
func (h *Hub) GetConnectedUsers(workspaceID string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if workspace, ok := h.workspaces[workspaceID]; ok {
		users := make([]string, 0, len(workspace))
		for userID := range workspace {
			users = append(users, userID)
		}
		return users
	}
	return []string{}
}

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

// readPump はWebSocketからのメッセージを読み取ります
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			_ = err // WebSocket接続のクローズエラーは無視
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return err
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] 予期しない切断エラー: user=%s workspace=%s error=%v",
					c.userID, c.workspaceID, err)
			} else {
				log.Printf("[WebSocket] 接続切断: user=%s workspace=%s error=%v", c.userID, c.workspaceID, err)
			}
			break
		}

		log.Printf("[WebSocket] メッセージ受信: user=%s workspace=%s サイズ=%d bytes",
			c.userID, c.workspaceID, len(message))

		// メッセージをパースして処理
		c.handleMessage(message)
	}
}

// handleMessage はクライアントからのメッセージを処理します
func (c *Client) handleMessage(data []byte) {
	var event chatv1.ClientEvent
	if err := protojson.Unmarshal(data, &event); err != nil {
		log.Printf("[WebSocket] メッセージパースエラー: user=%s error=%v", c.userID, err)
		c.sendError("PARSE_ERROR", "メッセージのパースに失敗しました")
		return
	}

	switch e := event.Event.(type) {
	case *chatv1.ClientEvent_JoinChannel:
		c.handleJoinChannel(e.JoinChannel.GetChannelId())
	case *chatv1.ClientEvent_LeaveChannel:
		c.handleLeaveChannel(e.LeaveChannel.GetChannelId())
	case *chatv1.ClientEvent_Typing:
		c.notifyTyping(e.Typing.GetChannelId(), true)
	case *chatv1.ClientEvent_StopTyping:
		c.notifyTyping(e.StopTyping.GetChannelId(), false)
	case *chatv1.ClientEvent_ViewChannel:
		c.handleViewChannel(e.ViewChannel.GetChannelId())
	default:
		log.Printf("[WebSocket] 未知のイベント: user=%s", c.userID)
		c.sendError("UNKNOWN_EVENT", "未知のイベントです")
	}
}

// handleJoinChannel はチャンネルの購読を開始します
func (c *Client) handleJoinChannel(channelID string) {
	if channelID == "" {
		c.sendError("INVALID_PAYLOAD", "無効なペイロードです")
		return
	}
	if !c.canAccessChannel(channelID) {
		c.sendError("FORBIDDEN", "チャンネルにアクセスできません")
		return
	}

	c.subscribedChannels[channelID] = true
	c.hub.subscribe <- &SubscribeRequest{
		WorkspaceID: c.workspaceID,
		ChannelID:   channelID,
		UserID:      c.userID,
	}

	log.Printf("[WebSocket] チャンネル購読追加: user=%s workspace=%s channel=%s 購読数=%d",
		c.userID, c.workspaceID, channelID, len(c.subscribedChannels))

	c.sendAck("join_channel")
}

// handleLeaveChannel はチャンネルの購読を解除します
func (c *Client) handleLeaveChannel(channelID string) {
	if channelID == "" {
		c.sendError("INVALID_PAYLOAD", "無効なペイロードです")
		return
	}

	delete(c.subscribedChannels, channelID)
	c.hub.unsubscribe <- &UnsubscribeRequest{
		WorkspaceID: c.workspaceID,
		ChannelID:   channelID,
		UserID:      c.userID,
	}

	log.Printf("[WebSocket] チャンネル購読解除: user=%s workspace=%s channel=%s 購読数=%d",
		c.userID, c.workspaceID, channelID, len(c.subscribedChannels))

	c.sendAck("leave_channel")
}

// handleViewChannel は閲覧中のチャンネルを更新します。空文字は閲覧をやめたことを表します
func (c *Client) handleViewChannel(channelID string) {
	if channelID != "" && !c.canAccessChannel(channelID) {
		c.sendError("FORBIDDEN", "チャンネルにアクセスできません")
		return
	}
	c.hub.SetViewingChannel(c, channelID)
	c.sendAck("view_channel")
}

// canAccessChannel は接続中のワークスペースのチャンネルで、閲覧権限があるかを確認します
func (c *Client) canAccessChannel(channelID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), writeWait)
	defer cancel()
	ch, err := c.channelAccess.EnsureChannelAccess(ctx, channelID, c.userID)
	if err != nil {
		log.Printf("[WebSocket] チャンネルへのアクセスを拒否しました: user=%s channel=%s error=%v", c.userID, channelID, err)
		return false
	}
	return ch.WorkspaceID == c.workspaceID
}

// sendEvent は接続中のクライアントにだけイベントを送信します
func (c *Client) sendEvent(event *chatv1.ServerEvent) {
	data := encodeServerEvent(event)
	if data == nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) sendAck(eventName string) {
	c.sendEvent(&chatv1.ServerEvent{Event: &chatv1.ServerEvent_Ack{Ack: &chatv1.AckEvent{Event: eventName, Success: true}}})
}

func (c *Client) sendError(code string, message string) {
	c.sendEvent(&chatv1.ServerEvent{Event: &chatv1.ServerEvent_Error{Error: &chatv1.ErrorEvent{Code: code, Message: message}}})
}

// writePump はWebSocketにメッセージを書き込みます
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			_ = err // WebSocket接続のクローズエラーは無視
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					_ = err // クローズメッセージの送信エラーは無視
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			// クライアントは1フレーム1イベントとして解釈するため結合しない
			if _, err := w.Write(message); err != nil {
				return
			}

			if err := w.Close(); err != nil {
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

// notifyTyping は入力中状態の開始・停止をチャンネルの他ユーザーに通知します
func (c *Client) notifyTyping(channelID string, typing bool) {
	if channelID == "" {
		c.sendError("INVALID_PAYLOAD", "無効なペイロードです")
		return
	}
	payload := &chatv1.TypingEvent{ChannelId: channelID, UserId: c.userID}
	event := &chatv1.ServerEvent{Event: &chatv1.ServerEvent_StopTyping{StopTyping: payload}}
	if typing {
		event.Event = &chatv1.ServerEvent_Typing{Typing: payload}
	}
	if data := encodeServerEvent(event); data != nil {
		c.hub.BroadcastToChannel(c.workspaceID, channelID, data, c.userID)
	}
}
