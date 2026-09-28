package http

import (
	"sync"
	"time"
)

// maxBuckets を超えたら満タンに戻ったバケットを捨ててメモリの増加を抑える
const maxBuckets = 10000

// rateLimiter はキーごとのトークンバケットです。プロセス内で完結するため、複数台構成では台数分まで許容されます
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

// allow は 1 回分を消費できれば true を、できなければ次に消費できるまでの時間を返します
func (l *rateLimiter) allow(key string) (bool, time.Duration) {
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
