package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

type Connection struct {
	hub         *Hub
	conn        *websocket.Conn
	send        chan []byte
	UserID      string
	WorkspaceID string
}

type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type ServerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewConnection(hub *Hub, conn *websocket.Conn, userID, workspaceID string) *Connection {
	return &Connection{
		hub:         hub,
		conn:        conn,
		send:        make(chan []byte, 256),
		UserID:      userID,
		WorkspaceID: workspaceID,
	}
}

func (c *Connection) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Parse and handle message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("json unmarshal error: %v", err)
			continue
		}

		// Handle different message types
		// This will be expanded with actual handlers
		log.Printf("received message type: %s from user: %s", clientMsg.Type, c.UserID)
	}
}

func (c *Connection) WritePump() {
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

			// Flush any additional messages
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

func (c *Connection) SendMessage(msgType string, payload interface{}) error {
	msg := ServerMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
	default:
		// Channel is full
	}
	return nil
}
