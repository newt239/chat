package websocket

import (
	"slices"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

// SetViewingChannel は接続が閲覧中のチャンネルを更新し、変化したチャンネルの閲覧者一覧を配信します
// 1 接続が閲覧できるチャンネルは 1 つだけで、空文字は閲覧していないことを表します
func (h *Hub) SetViewingChannel(client *Client, channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	previous := client.viewingChannel
	client.viewingChannel = channelID
	if previous != "" && previous != channelID {
		h.broadcastViewers(client.workspaceID, previous, nil)
	}
	if channelID != "" {
		// 購読の登録より先に届いた場合でも本人が一覧を受け取れるようにする
		h.broadcastViewers(client.workspaceID, channelID, client)
	}
}

// clearViewingChannel は切断した接続の閲覧状態を解除します。呼び出し側で h.mu をロックしてください
func (h *Hub) clearViewingChannel(client *Client) {
	if client.viewingChannel == "" {
		return
	}
	channelID := client.viewingChannel
	client.viewingChannel = ""
	h.broadcastViewers(client.workspaceID, channelID, nil)
}

// broadcastViewers はチャンネルの閲覧者一覧をチャンネル購読者と extra に送信します。呼び出し側で h.mu をロックしてください
func (h *Hub) broadcastViewers(workspaceID, channelID string, extra *Client) {
	workspace := h.workspaces[workspaceID]

	viewers := make([]string, 0)
	for userID, clients := range workspace {
		if slices.ContainsFunc(clients, func(c *Client) bool { return c.viewingChannel == channelID }) {
			viewers = append(viewers, userID)
		}
	}
	slices.Sort(viewers)

	data := encodeServerEvent(&chatv1.ServerEvent{Event: &chatv1.ServerEvent_ChannelViewers{
		ChannelViewers: &chatv1.ChannelViewersEvent{ChannelId: channelID, UserIds: viewers},
	}})
	if data == nil {
		return
	}

	sent := make(map[*Client]bool)
	for userID := range h.channelSubscribers[workspaceID][channelID] {
		for _, c := range workspace[userID] {
			sent[c] = true
			h.trySend(c, data)
		}
	}
	if extra != nil && !sent[extra] && slices.Contains(workspace[extra.userID], extra) {
		h.trySend(extra, data)
	}
}
