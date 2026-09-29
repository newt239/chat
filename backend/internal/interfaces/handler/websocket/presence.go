package websocket

import (
	"context"
	"log"
	"slices"
	"time"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
)

const (
	// 共有した閲覧の期限を延長する間隔。PresenceStore 側の期限より十分短くする
	presenceRefreshInterval = 30 * time.Second

	presenceTimeout = 2 * time.Second
)

// PresenceEntry は 1 接続が 1 チャンネルを閲覧していることを表します
type PresenceEntry struct {
	WorkspaceID string
	ChannelID   string
	ConnID      string
	UserID      string
}

// PresenceStore はチャンネルの閲覧者を全レプリカで共有します
// 延長されない閲覧は期限で消え、落ちたレプリカの接続が残り続けないようにします
type PresenceStore interface {
	Add(ctx context.Context, e PresenceEntry) error
	Remove(ctx context.Context, e PresenceEntry) error
	Refresh(ctx context.Context, entries []PresenceEntry) error
	// Viewers は閲覧中のユーザー ID を重複なしで昇順に返します
	Viewers(ctx context.Context, workspaceID, channelID string) ([]string, error)
}

// SetViewingChannel は接続が閲覧中のチャンネルを更新し、変化したチャンネルの閲覧者一覧を配信します
// 1 接続が閲覧できるチャンネルは 1 つだけで、空文字は閲覧していないことを表します
func (h *Hub) SetViewingChannel(client *Client, channelID string) {
	h.mu.Lock()
	if !h.isRegistered(client) {
		h.mu.Unlock()
		return
	}
	previous := client.viewingChannel
	client.viewingChannel = channelID
	h.mu.Unlock()

	if previous != "" && previous != channelID {
		h.leaveViewing(client, previous)
	}
	if channelID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), presenceTimeout)
	defer cancel()
	if h.presence != nil {
		if err := h.presence.Add(ctx, client.presenceEntry(channelID)); err != nil {
			log.Printf("[WebSocket] 閲覧者を登録できません: channel=%s error=%v", channelID, err)
		}
	}
	// 購読の登録より先に届いた場合でも本人が一覧を受け取れるようにする
	h.broadcastViewers(ctx, client.workspaceID, channelID, client)
}

// leaveViewing は接続がチャンネルの閲覧をやめたことを反映し、残った閲覧者を配信します
func (h *Hub) leaveViewing(client *Client, channelID string) {
	ctx, cancel := context.WithTimeout(context.Background(), presenceTimeout)
	defer cancel()
	if h.presence != nil {
		if err := h.presence.Remove(ctx, client.presenceEntry(channelID)); err != nil {
			log.Printf("[WebSocket] 閲覧者を削除できません: channel=%s error=%v", channelID, err)
		}
	}
	h.broadcastViewers(ctx, client.workspaceID, channelID, nil)
}

// broadcastViewers はチャンネルの閲覧者一覧をチャンネル購読者と extra に送信します
func (h *Hub) broadcastViewers(ctx context.Context, workspaceID, channelID string, extra *Client) {
	viewers, err := h.viewers(ctx, workspaceID, channelID)
	if err != nil {
		log.Printf("[WebSocket] 閲覧者を取得できません: channel=%s error=%v", channelID, err)
		return
	}
	data := encodeServerEvent(&chatv1.ServerEvent{Event: &chatv1.ServerEvent_ChannelViewers{
		ChannelViewers: &chatv1.ChannelViewersEvent{ChannelId: channelID, UserIds: viewers},
	}})
	if data == nil {
		return
	}

	h.BroadcastToChannelSubscribers(workspaceID, channelID, data)
	if extra == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.isRegistered(extra) && !h.isUserSubscribedToChannel(workspaceID, channelID, extra.userID) {
		h.trySend(extra, data)
	}
}

func (h *Hub) viewers(ctx context.Context, workspaceID, channelID string) ([]string, error) {
	if h.presence != nil {
		return h.presence.Viewers(ctx, workspaceID, channelID)
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	viewers := make([]string, 0)
	for userID, clients := range h.workspaces[workspaceID] {
		if slices.ContainsFunc(clients, func(c *Client) bool { return c.viewingChannel == channelID }) {
			viewers = append(viewers, userID)
		}
	}
	slices.Sort(viewers)
	return viewers, nil
}

// refreshPresence はこのプロセスの接続の閲覧を延長します
func (h *Hub) refreshPresence(ctx context.Context) {
	if h.presence == nil {
		return
	}
	var entries []PresenceEntry
	h.mu.RLock()
	for _, workspace := range h.workspaces {
		for _, clients := range workspace {
			for _, c := range clients {
				if c.viewingChannel != "" {
					entries = append(entries, c.presenceEntry(c.viewingChannel))
				}
			}
		}
	}
	h.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, presenceTimeout)
	defer cancel()
	if err := h.presence.Refresh(ctx, entries); err != nil {
		log.Printf("[WebSocket] 閲覧の期限を延長できません: error=%v", err)
	}
}

func (c *Client) presenceEntry(channelID string) PresenceEntry {
	return PresenceEntry{WorkspaceID: c.workspaceID, ChannelID: channelID, ConnID: c.id, UserID: c.userID}
}
