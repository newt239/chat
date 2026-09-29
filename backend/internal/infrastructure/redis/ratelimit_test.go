package redis

import (
	"testing"
	"time"
)

func TestRateLimiterIsSharedBetweenReplicas(t *testing.T) {
	ctx := t.Context()
	rdb := newTestClient(t)
	now := time.Unix(1000, 0)
	replicas := []*RateLimiter{NewRateLimiter(rdb, "webhook", 1, 2), NewRateLimiter(rdb, "webhook", 1, 2)}
	for _, l := range replicas {
		l.now = func() time.Time { return now }
	}

	for i, l := range replicas {
		if ok, _ := l.Allow(ctx, "a"); !ok {
			t.Fatalf("レプリカ %d の 1 回目が拒否されました", i)
		}
	}
	ok, wait := replicas[0].Allow(ctx, "a")
	if ok || wait != time.Second {
		t.Fatalf("全レプリカ合わせて上限を超えたら 1 秒待たせるはず: ok=%v wait=%v", ok, wait)
	}
	if ok, _ := replicas[1].Allow(ctx, "b"); !ok {
		t.Fatal("別のキーは制限されないはず")
	}

	now = now.Add(time.Second)
	if ok, _ := replicas[1].Allow(ctx, "a"); !ok {
		t.Fatal("時間が経てば再び受け付けるはず")
	}
}

func TestRateLimiterAllowsWhenRedisIsDown(t *testing.T) {
	rdb := newTestClient(t)
	_ = rdb.Close()
	if ok, _ := NewRateLimiter(rdb, "webhook", 1, 1).Allow(t.Context(), "a"); !ok {
		t.Fatal("Redis に届かないときは受け付けるはず")
	}
}
