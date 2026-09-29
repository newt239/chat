package http

import (
	"context"
	"sync"
	"time"
)

// Webhook ごとに毎秒 1 回、瞬間的には 10 回まで受け付ける
const (
	WebhookRatePerSecond = 1
	WebhookBurst         = 10
)

// RateLimiter はキーごとに受け付けてよいかを判定し、拒否するときは次に受け付けられるまでの時間を返します
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, time.Duration)
}

// maxBuckets を超えたら満タンに戻ったバケットを捨ててメモリの増加を抑える
const maxBuckets = 10000

// rateLimiter はプロセス内で完結するトークンバケットです。複数台構成では台数分まで許容されるため、Redis の実装を使います
type rateLimiter struct {
	mu      sync.Mutex
	rate    float64
	burst   float64
	buckets map[string]*bucket
	now     func() time.Time
}

type bucket struct {
	tokens  float64
	updated time.Time
}

func newRateLimiter(perSecond float64, burst int) *rateLimiter {
	return &rateLimiter{rate: perSecond, burst: float64(burst), buckets: map[string]*bucket{}, now: time.Now}
}

func (l *rateLimiter) Allow(_ context.Context, key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[key]
	if !ok {
		if len(l.buckets) >= maxBuckets {
			l.evictFull(now)
		}
		b = &bucket{tokens: l.burst, updated: now}
		l.buckets[key] = b
	}
	b.tokens = min(l.burst, b.tokens+now.Sub(b.updated).Seconds()*l.rate)
	b.updated = now
	if b.tokens < 1 {
		return false, time.Duration((1 - b.tokens) / l.rate * float64(time.Second))
	}
	b.tokens--
	return true, 0
}

func (l *rateLimiter) evictFull(now time.Time) {
	for key, b := range l.buckets {
		if b.tokens+now.Sub(b.updated).Seconds()*l.rate >= l.burst {
			delete(l.buckets, key)
		}
	}
}
