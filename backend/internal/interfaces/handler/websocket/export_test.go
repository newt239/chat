package websocket

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
