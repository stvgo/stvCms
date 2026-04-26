package clients

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -destination=../mocks/mock_redis_client.go -package=mocks stvCms/internal/clients IRedisClient

type IRedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

type redisWrapper struct {
	client *redis.Client
}

func NewRedisWrapper(client *redis.Client) IRedisClient {
	return &redisWrapper{client: client}
}

func (r *redisWrapper) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisWrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisWrapper) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
