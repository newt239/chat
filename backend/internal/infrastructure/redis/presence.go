package redis

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/newt239/chat/internal/domain/service"
)

// presenceTTL を過ぎても延長されない閲覧は、落ちたレプリカの接続とみなして消す
const presenceTTL = 90 * time.Second

// PresenceStore はチャンネルの閲覧者を全レプリカで共有します。チャンネルごとの sorted set に接続を期限付きで入れる
type PresenceStore struct {
	client *goredis.Client
	now    func() time.Time
}

func NewPresenceStore(client *goredis.Client) *PresenceStore {
	return &PresenceStore{client: client, now: time.Now}
}

func presenceKey(workspaceID, channelID string) string {
	return "chat:viewers:" + workspaceID + ":" + channelID
}

func presenceMember(e service.PresenceEntry) string {
	return e.ConnID + "|" + e.UserID
}

func (s *PresenceStore) Add(ctx context.Context, e service.PresenceEntry) error {
	return s.Refresh(ctx, []service.PresenceEntry{e})
}

func (s *PresenceStore) Refresh(ctx context.Context, entries []service.PresenceEntry) error {
	if len(entries) == 0 {
		return nil
	}
	expiresAt := float64(s.now().Add(presenceTTL).UnixMilli())
	pipe := s.client.Pipeline()
	for _, e := range entries {
		key := presenceKey(e.WorkspaceID, e.ChannelID)
		pipe.ZAdd(ctx, key, goredis.Z{Score: expiresAt, Member: presenceMember(e)})
		pipe.Expire(ctx, key, presenceTTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *PresenceStore) Remove(ctx context.Context, e service.PresenceEntry) error {
	return s.client.ZRem(ctx, presenceKey(e.WorkspaceID, e.ChannelID), presenceMember(e)).Err()
}

func (s *PresenceStore) Viewers(ctx context.Context, workspaceID, channelID string) ([]string, error) {
	key := presenceKey(workspaceID, channelID)
	pipe := s.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(s.now().UnixMilli(), 10))
	members := pipe.ZRange(ctx, key, 0, -1)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	viewers := make([]string, 0, len(members.Val()))
	for _, m := range members.Val() {
		if _, userID, ok := strings.Cut(m, "|"); ok {
			viewers = append(viewers, userID)
		}
	}
	slices.Sort(viewers)
	return slices.Compact(viewers), nil
}
