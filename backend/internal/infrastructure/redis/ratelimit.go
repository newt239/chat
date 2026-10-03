package redis

import (
	"context"
	"log/slog"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// tokenBucketScript はトークンバケットの補充と消費をアトミックに行い、{許可なら 1, 待つべきミリ秒} を返す
var tokenBucketScript = goredis.NewScript(`
local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local state = redis.call('HMGET', KEYS[1], 'tokens', 'updated')
local tokens = tonumber(state[1])
local updated = tonumber(state[2])
if tokens == nil or updated == nil then
  tokens = burst
  updated = now
end
tokens = math.min(burst, tokens + math.max(0, now - updated) / 1000 * rate)
local allowed = 0
local wait = 0
if tokens >= 1 then
  tokens = tokens - 1
  allowed = 1
else
  wait = math.ceil((1 - tokens) / rate * 1000)
end
redis.call('HSET', KEYS[1], 'tokens', tostring(tokens), 'updated', now)
redis.call('PEXPIRE', KEYS[1], math.ceil(burst / rate * 1000) + 1000)
return {allowed, wait}
`)

// RateLimiter は全レプリカで共有するキーごとのトークンバケットです
type RateLimiter struct {
	client *goredis.Client
	prefix string
	rate   float64
	burst  int
	now    func() time.Time
}

func NewRateLimiter(client *goredis.Client, prefix string, perSecond float64, burst int) *RateLimiter {
	return &RateLimiter{client: client, prefix: prefix, rate: perSecond, burst: burst, now: time.Now}
}

// Allow は Redis に届かないときは受け付け側に倒します。レート制限のためにサービスを止めない
func (l *RateLimiter) Allow(ctx context.Context, key string) (bool, time.Duration) {
	res, err := tokenBucketScript.Run(ctx, l.client, []string{"chat:ratelimit:" + l.prefix + ":" + key},
		l.rate, l.burst, l.now().UnixMilli()).Int64Slice()
	if err != nil || len(res) != 2 {
		slog.WarnContext(ctx, "レート制限を Redis で確認できないため受け付けます", "error", err)
		return true, 0
	}
	return res[0] == 1, time.Duration(res[1]) * time.Millisecond
}
