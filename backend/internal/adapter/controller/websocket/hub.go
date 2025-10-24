package websocket

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Hub はWebSocket接続を管理します
type Hub struct {
	// 登録されたクライアント
	clients map[*Client]bool

	// クライアントからの登録要求
	register chan *Client

	// クライアントからの登録解除要求
	unregister chan *Client

	// ブロードキャスト用のチャンネル
	broadcast chan []byte
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
}

// NewHub は新しいHubを作成します
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run はハブを開始します
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered: %s", client.userID)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s", client.userID)
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Broadcast は全クライアントにメッセージを送信します
func (h *Hub) Broadcast(message []byte) {
	select {
	case h.broadcast <- message:
	default:
	}
}

// BroadcastToUser は特定のユーザーにメッセージを送信します
func (h *Hub) BroadcastToUser(userID string, message []byte) {
	for client := range h.clients {
		if client.userID == userID {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
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
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// メッセージをハブにブロードキャスト
		c.hub.Broadcast(message)
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
