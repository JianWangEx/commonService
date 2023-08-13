// Package cache @Author  wangjian    2023/6/21 7:33 PM
package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// GetRedisClient
//
//	@Description: 返回一个使用redis的connection
//	@return redis.UniversalClient
func GetRedisClient() redis.UniversalClient {
	return client.redisClient
}

type Client interface {
	// Get receiver should be ptr and not nil
	Get(ctx context.Context, key string, receiver interface{}) error

	Set(ctx context.Context, key string, value interface{}, expired time.Duration) error

	Add(ctx context.Context, key string, value interface{}, expired time.Duration) error

	Delete(ctx context.Context, key string) error

	FlushCache(ctx context.Context) (string, error)
}
