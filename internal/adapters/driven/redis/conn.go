package redis

import (
	"context"
	"log"
	"marketstream/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, cfg config.RedisConfig) *Ouredis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password, // no password set
		DB:       cfg.DB,       // use default DB
	})
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	Ouredis := NewOuredis(client)
	log.Println("Redis connection established:", pong)
	return Ouredis
}
