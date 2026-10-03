package websocket

import (
	"context"
	"slices"
	"time"

	"go.uber.org/zap"

	"github.com/newt239/chat/internal/domain/service"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/infrastructure/logger"
)

const (
	// 共有した閲覧の期限を延長する間隔。PresenceStore 側の期限より十分短くする
	presenceRefreshInterval = 30 * time.Second

	presenceTimeout = 2 * time.Second
)

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
			logger.Get().Error("閲覧者を登録できません", zap.String("channel", channelID), zap.Error(err))
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
			logger.Get().Error("閲覧者を削除できません", zap.String("channel", channelID), zap.Error(err))
		}
	}
	h.broadcastViewers(ctx, client.workspaceID, channelID, nil)
}

// broadcastViewers はチャンネルの閲覧者一覧をチャンネル購読者と extra に送信します
func (h *Hub) broadcastViewers(ctx context.Context, workspaceID, channelID string, extra *Client) {
	viewers, err := h.viewers(ctx, workspaceID, channelID)
	if err != nil {
		logger.Get().Error("閲覧者を取得できません", zap.String("channel", channelID), zap.Error(err))
		return
	}
	data := encodeServerEvent(&chatv1.ServerEvent{Event: &chatv1.ServerEvent_ChannelViewers{
		ChannelViewers: &chatv1.ChannelViewersEvent{ChannelId: channelID, UserIds: viewers},
	}})
	if data == nil {
		return
	}

	h.BroadcastToChannel(workspaceID, channelID, data, "")
	if extra == nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if _, subscribed := extra.subscribedChannels[channelID]; h.isRegistered(extra) && !subscribed {
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
		for c := range clients {
			if c.viewingChannel == channelID {
				viewers = append(viewers, userID)
				break
			}
		}
	}
	slices.Sort(viewers)
	return viewers, nil
}

// runPresenceRefresh はこのプロセスの接続の閲覧を ctx が終わるまで定期的に延長します
func (h *Hub) runPresenceRefresh(ctx context.Context) {
	if h.presence == nil {
		<-ctx.Done()
		return
	}
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
	var entries []service.PresenceEntry
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
		logger.Get().Error("閲覧の期限を延長できません", zap.Error(err))
	}
}

func (c *Client) presenceEntry(channelID string) service.PresenceEntry {
	return service.PresenceEntry{WorkspaceID: c.workspaceID, ChannelID: channelID, ConnID: c.id, UserID: c.userID}
}
