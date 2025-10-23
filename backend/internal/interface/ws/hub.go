package ws

import (
	"sync"
)

type Hub struct {
	// workspace_id -> user_id -> connection
	workspaces map[string]map[string]*Connection
	mu         sync.RWMutex

	register   chan *Connection
	unregister chan *Connection
	broadcast  chan *BroadcastMessage
}

type BroadcastMessage struct {
	WorkspaceID string
	ChannelID   string
	ExcludeUser string
	Data        []byte
}

func NewHub() *Hub {
	return &Hub{
		workspaces: make(map[string]map[string]*Connection),
		register:   make(chan *Connection),
		unregister: make(chan *Connection),
		broadcast:  make(chan *BroadcastMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			if _, ok := h.workspaces[conn.WorkspaceID]; !ok {
				h.workspaces[conn.WorkspaceID] = make(map[string]*Connection)
			}
			h.workspaces[conn.WorkspaceID][conn.UserID] = conn
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			if users, ok := h.workspaces[conn.WorkspaceID]; ok {
				if _, ok := users[conn.UserID]; ok {
					delete(users, conn.UserID)
					close(conn.send)
					if len(users) == 0 {
						delete(h.workspaces, conn.WorkspaceID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if users, ok := h.workspaces[msg.WorkspaceID]; ok {
				for userID, conn := range users {
					if userID == msg.ExcludeUser {
						continue
					}
					select {
					case conn.send <- msg.Data:
					default:
						// Connection is blocked, close it
						close(conn.send)
						delete(users, userID)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

func (h *Hub) Broadcast(msg *BroadcastMessage) {
	h.broadcast <- msg
}

func (h *Hub) SendToUser(workspaceID, userID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if users, ok := h.workspaces[workspaceID]; ok {
		if conn, ok := users[userID]; ok {
			select {
			case conn.send <- data:
			default:
			}
		}
	}
}
