package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	log.Println("Redis connection established:", pong)
	return client
}
