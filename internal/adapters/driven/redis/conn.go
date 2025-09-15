package redisx

import (
	"context"
	"fmt"
	"marketstream/internal/config"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, cfg config.RedisConfig) (*goredis.Client, error) {
	const (
		maxRetries = 5
		retryDelay = 1 * time.Second
		pingTTL    = 2 * time.Second
	)

	var lastErr error
	for i := 1; i <= maxRetries; i++ {
		rdb := goredis.NewClient(&goredis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})

		pctx, cancel := context.WithTimeout(ctx, pingTTL)
		err := rdb.Ping(pctx).Err()
		cancel()

		if err == nil {
			return rdb, nil
		}
		_ = rdb.Close()
		lastErr = err

		select {
		case <-time.After(retryDelay):
		case <-ctx.Done():
			return nil, fmt.Errorf("redis connect canceled: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("redis unreachable after %d attempts: %w", maxRetries, lastErr)
}
