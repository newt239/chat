package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Hub はWebSocket接続を管理します
type Hub struct {
	// Workspace単位でクライアントを管理
	// workspaceID -> userID -> []*Client (同一ユーザーの複数接続をサポート)
	workspaces map[string]map[string][]*Client

	// クライアントからの登録要求
	register chan *Client

	// クライアントからの登録解除要求
	unregister chan *Client

	// ブロードキャスト用のチャンネル
	broadcast chan *BroadcastMessage
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
}

// NewHub は新しいHubを作成します
func NewHub() *Hub {
	return &Hub{
		workspaces: make(map[string]map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

// Run はハブを開始します
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
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

		case client := <-h.unregister:
			if workspace, ok := h.workspaces[client.workspaceID]; ok {
				if clients, ok := workspace[client.userID]; ok {
					// クライアントリストから削除
					for i, c := range clients {
						if c == client {
							workspace[client.userID] = append(clients[:i], clients[i+1:]...)
							close(client.send)
							break
						}
					}
					// クライアントがいなくなったらユーザーを削除
					if len(workspace[client.userID]) == 0 {
						delete(workspace, client.userID)
					}
					// Workspaceにユーザーがいなくなったら削除
					if len(workspace) == 0 {
						delete(h.workspaces, client.workspaceID)
					}
					log.Printf("[WebSocket] クライアント登録解除: user=%s workspace=%s 残接続数=%d",
						client.userID, client.workspaceID, len(workspace[client.userID]))
				}
			}

		case msg := <-h.broadcast:
			if workspace, ok := h.workspaces[msg.WorkspaceID]; ok {
				for userID, clients := range workspace {
					// ExcludeUserが設定されている場合はスキップ
					if msg.ExcludeUser != nil && userID == *msg.ExcludeUser {
						continue
					}
					// TODO: ChannelIDによるフィルタリングを実装
					for _, client := range clients {
						select {
						case client.send <- msg.Data:
						default:
							close(client.send)
						}
					}
				}
			}
		}
	}
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
func (h *Hub) BroadcastToChannel(workspaceID string, channelID string, message []byte) {
	h.broadcast <- &BroadcastMessage{
		WorkspaceID: workspaceID,
		ChannelID:   &channelID,
		Data:        message,
	}
	log.Printf("[WebSocket] Channelブロードキャスト: workspace=%s channel=%s サイズ=%d bytes",
		workspaceID, channelID, len(message))
}

// BroadcastToUser は特定のユーザーにメッセージを送信します
func (h *Hub) BroadcastToUser(workspaceID string, userID string, message []byte) {
	if workspace, ok := h.workspaces[workspaceID]; ok {
		if clients, ok := workspace[userID]; ok {
			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
				}
			}
			log.Printf("[WebSocket] ユーザー宛送信: workspace=%s user=%s 接続数=%d サイズ=%d bytes",
				workspaceID, userID, len(clients), len(message))
		}
	}
}

// GetConnectedUsers は指定されたWorkspace内の接続中のユーザーIDリストを返します
func (h *Hub) GetConnectedUsers(workspaceID string) []string {
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
	maxMessageSize = 512
)

// readPump はWebSocketからのメッセージを読み取ります
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WebSocket] 予期しない切断エラー: user=%s workspace=%s error=%v",
					c.userID, c.workspaceID, err)
			} else {
				log.Printf("[WebSocket] 接続切断: user=%s workspace=%s", c.userID, c.workspaceID)
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
	msg, err := ParseClientMessage(data)
	if err != nil {
		log.Printf("[WebSocket] メッセージパースエラー: user=%s error=%v", c.userID, err)
		c.sendError("PARSE_ERROR", "メッセージのパースに失敗しました")
		return
	}

	log.Printf("[WebSocket] イベント処理開始: type=%s user=%s workspace=%s",
		msg.Type, c.userID, c.workspaceID)

	// イベントタイプに応じた処理
	switch msg.Type {
	case EventTypeJoinChannel:
		c.handleJoinChannel(msg.Payload)
	case EventTypeLeaveChannel:
		c.handleLeaveChannel(msg.Payload)
	case EventTypePostMessage:
		c.handlePostMessage(msg.Payload)
	case EventTypeTyping:
		c.handleTyping(msg.Payload)
	case EventTypeUpdateReadState:
		c.handleUpdateReadState(msg.Payload)
	default:
		log.Printf("[WebSocket] 未知のイベントタイプ: type=%s user=%s", msg.Type, c.userID)
		c.sendError("UNKNOWN_EVENT", fmt.Sprintf("未知のイベントタイプです: %s", msg.Type))
	}
	log.Printf("[WebSocket] イベント処理完了: type=%s user=%s", msg.Type, c.userID)
}

// handleJoinChannel はjoin_channelイベントを処理します
func (c *Client) handleJoinChannel(payload json.RawMessage) {
	// TODO: チャンネル参加処理を実装
	log.Printf("User %s joining channel", c.userID)
	c.sendAck(EventTypeJoinChannel, true, "")
}

// handleLeaveChannel はleave_channelイベントを処理します
func (c *Client) handleLeaveChannel(payload json.RawMessage) {
	// TODO: チャンネル離脱処理を実装
	log.Printf("User %s leaving channel", c.userID)
	c.sendAck(EventTypeLeaveChannel, true, "")
}

// handlePostMessage はpost_messageイベントを処理します
func (c *Client) handlePostMessage(payload json.RawMessage) {
	// TODO: メッセージ投稿処理を実装（UseCase層との連携）
	log.Printf("User %s posting message", c.userID)
	c.sendAck(EventTypePostMessage, true, "")
}

// handleTyping はtypingイベントを処理します
func (c *Client) handleTyping(payload json.RawMessage) {
	// TODO: 入力中状態の通知処理を実装
	log.Printf("User %s is typing", c.userID)
}

// handleUpdateReadState はupdate_read_stateイベントを処理します
func (c *Client) handleUpdateReadState(payload json.RawMessage) {
	// TODO: 既読状態更新処理を実装
	log.Printf("User %s updating read state", c.userID)
	c.sendAck(EventTypeUpdateReadState, true, "")
}

// sendAck はACK応答を送信します
func (c *Client) sendAck(eventType EventType, success bool, message string) {
	payload := AckPayload{
		Type:    eventType,
		Success: success,
		Message: message,
	}
	data, err := SendServerMessage(EventTypeAck, payload)
	if err != nil {
		log.Printf("Failed to send ACK: %v", err)
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

// sendError はエラー応答を送信します
func (c *Client) sendError(code string, message string) {
	payload := ErrorPayload{
		Code:    code,
		Message: message,
	}
	data, err := SendServerMessage(EventTypeError, payload)
	if err != nil {
		log.Printf("Failed to send error: %v", err)
		return
	}
	select {
	case c.send <- data:
	default:
	}
}

// writePump はWebSocketにメッセージを書き込みます
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// キューされたメッセージを追加
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
