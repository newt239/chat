package websocket

import (
	"context"
	"log/slog"
	"time"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/infrastructure/redis"
)

const (
	// 共有した閲覧の期限を延長する間隔。PresenceStore 側の期限より十分短くする
	presenceRefreshInterval = 30 * time.Second

	presenceTimeout = 2 * time.Second
)

// SetViewingChannel は接続が閲覧中のチャンネル (1 つだけ、空文字はなし) を更新し、変化したチャンネルの閲覧者一覧を配信します
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
	if err := h.presence.Add(ctx, client.presenceEntry(channelID)); err != nil {
		slog.Error("閲覧者を登録できません", "channel", channelID, "error", err)
	}
	// 購読の登録より先に届いた場合でも本人が一覧を受け取れるようにする
	h.broadcastViewers(ctx, client.workspaceID, channelID, client)
}

// leaveViewing は接続がチャンネルの閲覧をやめたことを反映し、残った閲覧者を配信します
func (h *Hub) leaveViewing(client *Client, channelID string) {
	ctx, cancel := context.WithTimeout(context.Background(), presenceTimeout)
	defer cancel()
	if err := h.presence.Remove(ctx, client.presenceEntry(channelID)); err != nil {
		slog.Error("閲覧者を削除できません", "channel", channelID, "error", err)
	}
	h.broadcastViewers(ctx, client.workspaceID, channelID, nil)
}

// broadcastViewers はチャンネルの閲覧者一覧をチャンネル購読者と extra に送信します
func (h *Hub) broadcastViewers(ctx context.Context, workspaceID, channelID string, extra *Client) {
	viewers, err := h.presence.Viewers(ctx, workspaceID, channelID)
	if err != nil {
		slog.Error("閲覧者を取得できません", "channel", channelID, "error", err)
		return
	}
	data := encodeServerEvent(&chatv1.ServerEvent{Event: &chatv1.ServerEvent_ChannelViewers{
		ChannelViewers: &chatv1.ChannelViewersEvent{ChannelId: channelID, UserIds: viewers},
	}})
	if data == nil {
		return
	}

	h.publish(envelope{Target: targetChannel, WorkspaceID: workspaceID, ChannelID: channelID, Data: data})
	if extra == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if _, subscribed := extra.subscribedChannels[channelID]; h.isRegistered(extra) && !subscribed {
		h.trySend(extra, data)
	}
}

// runPresenceRefresh はこのプロセスの接続の閲覧を ctx が終わるまで定期的に延長します
func (h *Hub) runPresenceRefresh(ctx context.Context) {
	ticker := time.NewTicker(presenceRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.refreshPresence(ctx)
		}
	}
}

func (h *Hub) refreshPresence(ctx context.Context) {
	var entries []redis.PresenceEntry
	h.mu.RLock()
	for _, users := range h.workspaces {
		for _, clients := range users {
			for c := range clients {
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
		slog.Error("閲覧の期限を延長できません", "error", err)
	}
}

func (c *Client) presenceEntry(channelID string) redis.PresenceEntry {
	return redis.PresenceEntry{WorkspaceID: c.workspaceID, ChannelID: channelID, ConnID: c.id, UserID: c.userID}
}
