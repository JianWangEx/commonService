// Package cache @Author  wangjian    2023/6/22 5:46 PM
package cache

import (
	"context"
	"github.com/patrickmn/go-cache"
	"time"
)

type LocalCacheManager struct {
	Cache *cache.Cache
}

func (c *LocalCacheManager) Get(ctx context.Context, key string) (interface{}, bool) {
	return c.Cache.Get(key)
}

func (c *LocalCacheManager) Set(ctx context.Context, key string, value interface{}, d time.Duration) {
	c.Cache.Set(key, value, d)
}

func (c *LocalCacheManager) Delete(ctx context.Context, key string) {
	c.Cache.Delete(key)
}

func (c *LocalCacheManager) Add(ctx context.Context, key string, value interface{}, d time.Duration) error {
	return c.Cache.Add(key, value, d)
}
