package clients

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, addr, password string) *redis.Client {
	var rdb *redis.Client

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("REDIS_URL inválida: %v", err)
		}
		rdb = redis.NewClient(opt)
	} else {
		if addr == "" {
			addr = "localhost:6379"
		}
		rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		})
	}

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("no se pudo conectar a Redis: %v", err)
	}

	return rdb
}
