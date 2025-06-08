package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Ouredis struct {
	red *redis.Client
}

func NewOuredis(red *redis.Client) *Ouredis {
	return &Ouredis{
		red: red,
	}
}

func (ouredis *Ouredis) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return ouredis.red.Set(ctx, key, value, duration).Err()
}
