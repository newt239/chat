package websocket

// NewTestClient は接続を持たないクライアントをハブに登録し、指定したチャンネルを購読させます
func NewTestClient(h *Hub, workspaceID, userID string, channels ...string) *Client {
	c := &Client{hub: h, id: userID + "@" + workspaceID, send: make(chan []byte, 32), userID: userID, workspaceID: workspaceID}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.workspaces[workspaceID] == nil {
		h.workspaces[workspaceID] = map[string][]*Client{}
	}
	h.workspaces[workspaceID][userID] = append(h.workspaces[workspaceID][userID], c)
	for _, ch := range channels {
		if h.channelSubscribers[workspaceID] == nil {
			h.channelSubscribers[workspaceID] = map[string]map[string]bool{}
		}
		if h.channelSubscribers[workspaceID][ch] == nil {
			h.channelSubscribers[workspaceID][ch] = map[string]bool{}
		}
		h.channelSubscribers[workspaceID][ch][userID] = true
	}
	return c
}

// Sent はクライアントに送られたデータを返します
func (c *Client) Sent() <-chan []byte {
	return c.send
}
