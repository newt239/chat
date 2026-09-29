package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// NewClient は REDIS_URL から接続し、応答するまで待ちます
func NewClient(url string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse REDIS_URL: %w", err)
	}
	client := goredis.NewClient(opts)

	const maxRetries = 10
	for i := range maxRetries {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = client.Ping(ctx).Err()
		cancel()
		if err == nil {
			return client, nil
		}
		if i < maxRetries-1 {
			time.Sleep(time.Second)
		}
	}
	_ = client.Close()
	return nil, fmt.Errorf("redis is not reachable: %w", err)
}
