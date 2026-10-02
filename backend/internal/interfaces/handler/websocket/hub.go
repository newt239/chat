package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/infrastructure/logger"
)

// Broker はイベントを全レプリカに配ります。Subscribe は ctx が終わるまで受信を続けます
type Broker interface {
	Publish(ctx context.Context, payload []byte) error
	Subscribe(ctx context.Context, handle func([]byte)) error
}

// HubOption は Hub の配信とチャンネル閲覧者の共有先を切り替えます。指定しなければプロセス内で完結します
type HubOption func(*Hub)

func WithBroker(b Broker) HubOption {
	return func(h *Hub) { h.broker = b }
}

func WithPresenceStore(p service.PresenceStore) HubOption {
	return func(h *Hub) { h.presence = p }
}

// アプリ独自の切断コード。クライアントは 4401 ならトークンを更新し、4403 ならワークスペースから離れる
const (
	CloseUnauthenticated = 4401
	CloseForbidden       = 4403
)

const (
	// 他のレプリカへの配信を待つ上限
	publishTimeout = 2 * time.Second
	// 他のレプリカへの配信を待たせておける件数。溢れたらこのプロセスの接続にだけ届ける
	outboxSize = 1024
)

type clientSet map[*Client]struct{}

// Hub はWebSocket接続を管理します
type Hub struct {
	// workspaces・subscribers と各接続の subscribedChannels・viewingChannel を保護する
	mu sync.RWMutex
	// workspaceID -> userID -> 接続。同じユーザーが複数の端末やタブから接続できる
	workspaces map[string]map[string]clientSet
	// workspaceID -> channelID -> 購読している接続
	subscribers map[string]map[string]clientSet

	// 書き込みを終えるまで Shutdown で待つ
	pumps sync.WaitGroup

	channelAccess service.ChannelAccessService
	// nil ならこのプロセスの接続にだけ配信する
	broker Broker
	outbox chan envelope
	// nil ならこのプロセスの接続から閲覧者を数える
	presence service.PresenceStore
}

type target string

const (
	targetWorkspace target = "workspace"
	targetChannel   target = "channel"
	targetUsers     target = "users"
	// 購読を外す。UserID が空なら購読者ごとに閲覧権限を確かめ直す
	targetRevokeChannel target = "revokeChannel"
	// ワークスペースから外されたユーザーの接続を 4403 で切る
	targetCloseWorkspaceUser target = "closeWorkspaceUser"
	// 失効したセッションの接続を 4401 で切る
	targetCloseSession target = "closeSession"
	// ユーザーの全接続を 4401 で切る
	targetCloseUser target = "closeUser"
)

// envelope はレプリカ間で受け渡す配信内容です。Data はエンコード済みの ServerEvent
type envelope struct {
	Target        target          `json:"target"`
	WorkspaceID   string          `json:"workspaceId,omitempty"`
	ChannelID     string          `json:"channelId,omitempty"`
	UserID        string          `json:"userId,omitempty"`
	UserIDs       []string        `json:"userIds,omitempty"`
	SessionID     string          `json:"sessionId,omitempty"`
	ExcludeUserID string          `json:"excludeUserId,omitempty"`
	Data          json.RawMessage `json:"data,omitempty"`
}

// NewHub は新しいHubを作成します
func NewHub(channelAccess service.ChannelAccessService, opts ...HubOption) *Hub {
	h := &Hub{
		workspaces:    make(map[string]map[string]clientSet),
		subscribers:   make(map[string]map[string]clientSet),
		channelAccess: channelAccess,
		outbox:        make(chan envelope, outboxSize),
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Run は他のレプリカとの送受信と閲覧の延長を、ctx が終わるまで続けます
func (h *Hub) Run(ctx context.Context) {
	if h.broker != nil {
		go h.runSubscriber(ctx)
		go h.runPublisher(ctx)
	}
	h.runPresenceRefresh(ctx)
}

func (h *Hub) register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	users := h.workspaces[c.workspaceID]
	if users == nil {
		users = make(map[string]clientSet)
		h.workspaces[c.workspaceID] = users
	}
	if users[c.userID] == nil {
		users[c.userID] = make(clientSet)
	}
	users[c.userID][c] = struct{}{}
}

// unregister は接続をハブから外し、closeMessage を送って切断させます。登録済みでなければ何もしません
func (h *Hub) unregister(c *Client, closeMessage []byte) {
	h.mu.Lock()
	users := h.workspaces[c.workspaceID]
	if _, ok := users[c.userID][c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(users[c.userID], c)
	if len(users[c.userID]) == 0 {
		delete(users, c.userID)
	}
	if len(users) == 0 {
		delete(h.workspaces, c.workspaceID)
	}
	for channelID := range c.subscribedChannels {
		h.removeSubscriber(c, channelID)
	}
	viewing := c.viewingChannel
	c.viewingChannel = ""
	c.closeMessage = closeMessage
	close(c.send)
	h.mu.Unlock()

	if viewing != "" {
		h.leaveViewing(c, viewing)
	}
}

// disconnect はアプリ独自の切断コードで接続を切ります
func (h *Hub) disconnect(c *Client, code int, reason string) {
	h.unregister(c, websocket.FormatCloseMessage(code, reason))
}

func (h *Hub) subscribe(c *Client, channelID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.isRegistered(c) {
		return false
	}
	channels := h.subscribers[c.workspaceID]
	if channels == nil {
		channels = make(map[string]clientSet)
		h.subscribers[c.workspaceID] = channels
	}
	if channels[channelID] == nil {
		channels[channelID] = make(clientSet)
	}
	channels[channelID][c] = struct{}{}
	c.subscribedChannels[channelID] = struct{}{}
	return true
}

func (h *Hub) unsubscribe(c *Client, channelID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeSubscriber(c, channelID)
}

// removeSubscriber は h.mu をロックした状態で呼んでください
func (h *Hub) removeSubscriber(c *Client, channelID string) {
	delete(c.subscribedChannels, channelID)
	channels := h.subscribers[c.workspaceID]
	delete(channels[channelID], c)
	if len(channels[channelID]) == 0 {
		delete(channels, channelID)
	}
	if len(channels) == 0 {
		delete(h.subscribers, c.workspaceID)
	}
}

func (h *Hub) isSubscribed(c *Client, channelID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := c.subscribedChannels[channelID]
	return ok
}

// isRegistered は接続がまだハブに登録されているかを返します。呼び出し側で h.mu をロックしてください
func (h *Hub) isRegistered(c *Client) bool {
	_, ok := h.workspaces[c.workspaceID][c.userID][c]
	return ok
}

// canAccess は接続中のワークスペースのチャンネルで、閲覧権限があるかを確認します
func (h *Hub) canAccess(c *Client, channelID string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), writeWait)
	defer cancel()
	ch, err := h.channelAccess.EnsureChannelAccess(ctx, channelID, c.userID)
	if err != nil {
		logger.Get().Debug("チャンネルへのアクセスを拒否しました", zap.String("user", c.userID), zap.String("channel", channelID), zap.Error(err))
		return false
	}
	return ch.WorkspaceID == c.workspaceID
}

// runSubscriber は他のレプリカが送ったイベントを受け取り、このプロセスの接続に配信します
func (h *Hub) runSubscriber(ctx context.Context) {
	for {
		err := h.broker.Subscribe(ctx, func(payload []byte) {
			var env envelope
			if err := json.Unmarshal(payload, &env); err != nil {
				logger.Get().Error("配信内容を読めません", zap.Error(err))
				return
			}
			h.deliver(&env)
		})
		if ctx.Err() != nil {
			return
		}
		logger.Get().Warn("イベントの購読が切れたため再開します", zap.Error(err))
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

// runPublisher は配信を待たせずに済むよう、他のレプリカへの送信を 1 つのワーカーで順に行います
func (h *Hub) runPublisher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case env := <-h.outbox:
			payload, err := json.Marshal(env)
			if err == nil {
				pubCtx, cancel := context.WithTimeout(ctx, publishTimeout)
				err = h.broker.Publish(pubCtx, payload)
				cancel()
			}
			if err != nil {
				// 送れなかったときはせめてこのプロセスの接続には届ける
				logger.Get().Error("他のレプリカへの配信に失敗しました", zap.String("target", string(env.Target)), zap.Error(err))
				h.deliver(&env)
			}
		}
	}
}

// publish は全レプリカに配信します
func (h *Hub) publish(env envelope) {
	if h.broker == nil {
		h.deliver(&env)
		return
	}
	select {
	case h.outbox <- env:
	default:
		logger.Get().Warn("他のレプリカへの配信が詰まっているため、このプロセスの接続にだけ届けます", zap.String("target", string(env.Target)))
		h.deliver(&env)
	}
}

// deliver はこのプロセスが持つ接続のうち、配信先に当たるものへ送信します
func (h *Hub) deliver(env *envelope) {
	switch env.Target {
	case targetRevokeChannel:
		// 閲覧権限の確認で DB を引くため、配信の受信を止めない
		go h.revokeChannel(env.WorkspaceID, env.ChannelID, env.UserID)
		return
	case targetCloseWorkspaceUser:
		h.disconnectMatching(CloseForbidden, "removed from workspace", func(c *Client) bool {
			return c.workspaceID == env.WorkspaceID && c.userID == env.UserID
		})
		return
	case targetCloseSession:
		h.disconnectMatching(CloseUnauthenticated, "session revoked", func(c *Client) bool { return c.sessionID == env.SessionID })
		return
	case targetCloseUser:
		h.disconnectMatching(CloseUnauthenticated, "session revoked", func(c *Client) bool { return c.userID == env.UserID })
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	data := []byte(env.Data)
	users := h.workspaces[env.WorkspaceID]
	switch env.Target {
	case targetWorkspace:
		for _, clients := range users {
			for c := range clients {
				h.trySend(c, data)
			}
		}
	case targetChannel:
		for c := range h.subscribers[env.WorkspaceID][env.ChannelID] {
			if c.userID != env.ExcludeUserID {
				h.trySend(c, data)
			}
		}
	case targetUsers:
		for _, userID := range env.UserIDs {
			for c := range users[userID] {
				h.trySend(c, data)
			}
		}
	}
}

// trySend は送信バッファが詰まっている接続を切ります。h.mu を持った状態で呼ぶため、切断は別の goroutine で行う
func (h *Hub) trySend(c *Client, data []byte) {
	select {
	case c.send <- data:
	default:
		go h.disconnect(c, websocket.CloseTryAgainLater, "too slow")
	}
}

func (h *Hub) matching(match func(*Client) bool) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var result []*Client
	for _, users := range h.workspaces {
		for _, clients := range users {
			for c := range clients {
				if match(c) {
					result = append(result, c)
				}
			}
		}
	}
	return result
}

func (h *Hub) disconnectMatching(code int, reason string, match func(*Client) bool) {
	for _, c := range h.matching(match) {
		h.disconnect(c, code, reason)
	}
}

// revokeChannel はチャンネルを見られなくなった接続の購読と閲覧をやめさせます
func (h *Hub) revokeChannel(workspaceID, channelID, userID string) {
	h.mu.RLock()
	var targets []*Client
	for c := range h.subscribers[workspaceID][channelID] {
		if userID == "" || c.userID == userID {
			targets = append(targets, c)
		}
	}
	for _, clients := range h.workspaces[workspaceID] {
		for c := range clients {
			if c.viewingChannel == channelID && (userID == "" || c.userID == userID) {
				targets = append(targets, c)
			}
		}
	}
	h.mu.RUnlock()

	checked := make(map[*Client]bool, len(targets))
	for _, c := range targets {
		if _, ok := checked[c]; ok {
			continue
		}
		revoke := userID != "" || !h.canAccess(c, channelID)
		checked[c] = revoke
		if !revoke {
			continue
		}
		h.unsubscribe(c, channelID)
		h.mu.RLock()
		viewing := c.viewingChannel == channelID
		h.mu.RUnlock()
		if viewing {
			h.SetViewingChannel(c, "")
		}
	}
}

// BroadcastToWorkspace はWorkspace内の全クライアントにメッセージを送信します
func (h *Hub) BroadcastToWorkspace(workspaceID string, message []byte) {
	h.publish(envelope{Target: targetWorkspace, WorkspaceID: workspaceID, Data: message})
}

// BroadcastToChannel はチャンネルを購読している接続に送信します。excludeUserID が空でなければそのユーザーには送りません
func (h *Hub) BroadcastToChannel(workspaceID string, channelID string, message []byte, excludeUserID string) {
	h.publish(envelope{Target: targetChannel, WorkspaceID: workspaceID, ChannelID: channelID, ExcludeUserID: excludeUserID, Data: message})
}

// BroadcastToUsers は指定したユーザーの全接続に送信します
func (h *Hub) BroadcastToUsers(workspaceID string, userIDs []string, message []byte) {
	h.publish(envelope{Target: targetUsers, WorkspaceID: workspaceID, UserIDs: userIDs, Data: message})
}

// RevokeChannel はユーザーのチャンネルの購読を外します。userID が空なら購読者ごとに閲覧権限を確かめ直します
func (h *Hub) RevokeChannel(workspaceID, channelID, userID string) {
	h.publish(envelope{Target: targetRevokeChannel, WorkspaceID: workspaceID, ChannelID: channelID, UserID: userID})
}

// CloseWorkspaceUser はワークスペースから外されたユーザーの接続を切ります
func (h *Hub) CloseWorkspaceUser(workspaceID, userID string) {
	h.publish(envelope{Target: targetCloseWorkspaceUser, WorkspaceID: workspaceID, UserID: userID})
}

// CloseSession は失効したセッションの接続を切ります
func (h *Hub) CloseSession(sessionID string) {
	h.publish(envelope{Target: targetCloseSession, SessionID: sessionID})
}

// CloseUser はユーザーの全接続を切ります
func (h *Hub) CloseUser(userID string) {
	h.publish(envelope{Target: targetCloseUser, UserID: userID})
}

// Shutdown は全接続に Going Away の close フレームを送り、クライアントが他のレプリカへつなぎ直すのを促します
// ctx が終わるまでに送り終えなかった接続は待たずに戻ります
func (h *Hub) Shutdown(ctx context.Context) {
	clients := h.clients()
	for _, c := range clients {
		h.disconnect(c, websocket.CloseGoingAway, "server shutting down")
	}
	logger.Get().Info("停止のため接続に切断を通知しました", zap.Int("connections", len(clients)))

	done := make(chan struct{})
	go func() {
		h.pumps.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (h *Hub) clients() []*Client {
	return h.matching(func(*Client) bool { return true })
}
