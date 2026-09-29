package websocket

import (
	"context"
	"encoding/json"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/newt239/chat/internal/domain/service"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

// Broker はイベントを全レプリカに配ります。Subscribe は ctx が終わるまで受信を続けます
type Broker interface {
	Publish(ctx context.Context, payload []byte) error
	Subscribe(ctx context.Context, handle func([]byte)) error
}

// HubOption は Hub の配信とチャンネル閲覧者の共有先を切り替えます。指定しなければプロセス内で完結します
type HubOption func(*Hub)

func WithBroker(b Broker) HubOption {
	return func(h *Hub) { h.broker = b }
}

func WithPresenceStore(p PresenceStore) HubOption {
	return func(h *Hub) { h.presence = p }
}

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

	// チャンネル購読管理用チャンネル
	subscribe   chan *SubscribeRequest
	unsubscribe chan *UnsubscribeRequest

	// nil ならこのプロセスの接続にだけ配信する
	broker Broker
	// nil ならこのプロセスの接続から閲覧者を数える
	presence PresenceStore
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

type target string

const (
	targetWorkspace target = "workspace"
	targetChannel   target = "channel"
	targetUser      target = "user"
)

// envelope はレプリカ間で受け渡す配信内容です。Data はエンコード済みの ServerEvent
type envelope struct {
	Target        target          `json:"target"`
	WorkspaceID   string          `json:"workspaceId"`
	ChannelID     string          `json:"channelId,omitempty"`
	UserID        string          `json:"userId,omitempty"`
	ExcludeUserID string          `json:"excludeUserId,omitempty"`
	Data          json.RawMessage `json:"data"`
}

// Client はWebSocket接続を表します
type Client struct {
	// WebSocketハブ
	hub *Hub

	// WebSocket接続
	conn *websocket.Conn

	// 閲覧者の共有で接続を区別するための ID
	id string

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
func NewHub(opts ...HubOption) *Hub {
	h := &Hub{
		workspaces:         make(map[string]map[string][]*Client),
		channelSubscribers: make(map[string]map[string]map[string]bool),
		register:           make(chan *Client),
		unregister:         make(chan *Client, 256),
		subscribe:          make(chan *SubscribeRequest),
		unsubscribe:        make(chan *UnsubscribeRequest),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Run はハブを開始します。ctx が終わると他のレプリカからの受信と閲覧の延長をやめます
func (h *Hub) Run(ctx context.Context) {
	if h.broker != nil {
		go h.runSubscriber(ctx)
	}
	refresh := time.NewTicker(presenceRefreshInterval)
	defer refresh.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-refresh.C:
			h.refreshPresence(ctx)

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
			viewing := ""
			if workspace, ok := h.workspaces[client.workspaceID]; ok {
				if clients, ok := workspace[client.userID]; ok {
					// クライアントリストから削除
					for i, c := range clients {
						if c == client {
							workspace[client.userID] = append(clients[:i], clients[i+1:]...)
							close(client.send)
							viewing = client.viewingChannel
							client.viewingChannel = ""
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
			if viewing != "" {
				h.leaveViewing(client, viewing)
			}

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
		}
	}
}

// runSubscriber は他のレプリカが送ったイベントを受け取り、このプロセスの接続に配信します
func (h *Hub) runSubscriber(ctx context.Context) {
	for {
		err := h.broker.Subscribe(ctx, func(payload []byte) {
			var env envelope
			if err := json.Unmarshal(payload, &env); err != nil {
				log.Printf("[WebSocket] 配信内容を読めません: %v", err)
				return
			}
			h.deliver(&env)
		})
		if ctx.Err() != nil {
			return
		}
		log.Printf("[WebSocket] イベントの購読が切れたため再開します: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

// publish は全レプリカに配信します。送れなかったときはせめてこのプロセスの接続には届ける
func (h *Hub) publish(env envelope) {
	if h.broker == nil {
		h.deliver(&env)
		return
	}
	payload, err := json.Marshal(env)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
		err = h.broker.Publish(ctx, payload)
		cancel()
		if err == nil {
			return
		}
	}
	log.Printf("[WebSocket] 他のレプリカへの配信に失敗しました: target=%s workspace=%s error=%v", env.Target, env.WorkspaceID, err)
	h.deliver(&env)
}

// deliver はこのプロセスが持つ接続のうち、配信先に当たるものへ送信します
func (h *Hub) deliver(env *envelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	workspace, ok := h.workspaces[env.WorkspaceID]
	if !ok {
		return
	}
	data := []byte(env.Data)
	switch env.Target {
	case targetWorkspace:
		for _, clients := range workspace {
			for _, client := range clients {
				h.trySend(client, data)
			}
		}
	case targetChannel:
		for userID := range h.channelSubscribers[env.WorkspaceID][env.ChannelID] {
			if userID == env.ExcludeUserID {
				continue
			}
			for _, client := range workspace[userID] {
				h.trySend(client, data)
			}
		}
	case targetUser:
		for _, client := range workspace[env.UserID] {
			h.trySend(client, data)
		}
	}
}

// trySend は送信バッファが詰まっている接続を切断対象にします。close と競合しないよう h.mu を持った状態で呼んでください
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

// isRegistered は接続がまだハブに登録されているかを返します。呼び出し側で h.mu をロックしてください
func (h *Hub) isRegistered(client *Client) bool {
	return slices.Contains(h.workspaces[client.workspaceID][client.userID], client)
}

// BroadcastToWorkspace はWorkspace内の全クライアントにメッセージを送信します
func (h *Hub) BroadcastToWorkspace(workspaceID string, message []byte) {
	h.publish(envelope{Target: targetWorkspace, WorkspaceID: workspaceID, Data: message})
}

// BroadcastToChannel はChannel内の全クライアントにメッセージを送信します
// excludeUserID が空でない場合はそのユーザーを配信対象から除外します
func (h *Hub) BroadcastToChannel(workspaceID string, channelID string, message []byte, excludeUserID string) {
	h.publish(envelope{Target: targetChannel, WorkspaceID: workspaceID, ChannelID: channelID, ExcludeUserID: excludeUserID, Data: message})
}

// BroadcastToUser は特定のユーザーにメッセージを送信します
func (h *Hub) BroadcastToUser(workspaceID string, userID string, message []byte) {
	h.publish(envelope{Target: targetUser, WorkspaceID: workspaceID, UserID: userID, Data: message})
}

// BroadcastToChannelSubscribers はチャンネルを購読している全ユーザーにメッセージを送信します
// メッセージイベント(新着/編集/削除)の配信に使用します
func (h *Hub) BroadcastToChannelSubscribers(workspaceID string, channelID string, message []byte) {
	h.BroadcastToChannel(workspaceID, channelID, message, "")
}

// Shutdown は全接続に Going Away の close フレームを送り、クライアントが他のレプリカへつなぎ直すのを促します
// ctx が終わるまでに切断が済まなかった接続は強制的に閉じます
func (h *Hub) Shutdown(ctx context.Context) {
	clients := h.clients()
	closeMessage := websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down")
	for _, c := range clients {
		_ = c.conn.WriteControl(websocket.CloseMessage, closeMessage, time.Now().Add(writeWait))
	}
	log.Printf("[WebSocket] 停止のため %d 接続に切断を通知しました", len(clients))

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for len(h.clients()) > 0 {
		select {
		case <-ctx.Done():
			for _, c := range h.clients() {
				_ = c.conn.Close()
			}
			return
		case <-ticker.C:
		}
	}
}

func (h *Hub) clients() []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var all []*Client
	for _, workspace := range h.workspaces {
		for _, clients := range workspace {
			all = append(all, clients...)
		}
	}
	return all
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

	// 他のレプリカへの配信を待つ上限
	publishTimeout = 2 * time.Second
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
