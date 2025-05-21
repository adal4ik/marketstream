package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"

	"marketstream/internal/config"
)

func NewRedis(ctx context.Context, cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password, // no password set
		DB:       cfg.DB,       // use default DB
	})
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	log.Println("Redis connection established:", pong)
	return client
}
