// Package cache @Author  wangjian    2023/6/29 11:45 AM
package cache

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type Handler interface {
	GetLocalCacheKey(ctx context.Context) (string, error)
	GetRedisCacheKey(ctx context.Context) (string, error)
	SetLocalCache(ctx context.Context, key string, value interface{}, expired time.Duration) error
	SetRedisCache(ctx context.Context, key string, value interface{}, expired time.Duration) error
	AddLocalCache(ctx context.Context, key string, value interface{}, expired time.Duration) error
	AddRedisCache(ctx context.Context, key string, value interface{}, expired time.Duration) error
	GetLocalCache(ctx context.Context, key string, receiver interface{}) error
	GetRedisCache(ctx context.Context, key string, receiver interface{}) error
	DeleteLocalCache(ctx context.Context, key string) error
	DeleteRedisCache(ctx context.Context, key string) error
}

type BaseHandler struct {
	client *cacheManager
}

var handler *BaseHandler

func GetBaseHandler() *BaseHandler {
	return handler
}

func (h *BaseHandler) GetLocalCacheKey(ctx context.Context) (string, error) {
	return "", errors.New("GetLocalCacheKey is not implemented")
}

func (h *BaseHandler) GetRedisCacheKey(ctx context.Context) (string, error) {
	return "", errors.New("GetRedisCacheKey is not implemented")
}

func (h *BaseHandler) SetLocalCache(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	return h.client.Set(ctx, key, value, expired)
}

func (h *BaseHandler) SetRedisCache(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	return h.client.Set(ctx, key, value, expired)
}

func (h *BaseHandler) AddLocalCache(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	return h.client.Add(ctx, key, value, expired)
}

func (h *BaseHandler) AddRedisCache(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	return h.client.Add(ctx, key, value, expired)
}

func (h *BaseHandler) GetLocalCache(ctx context.Context, key string, receiver interface{}) error {
	return h.client.Get(ctx, key, receiver)
}

func (h *BaseHandler) GetRedisCache(ctx context.Context, key string, receiver interface{}) error {
	return h.client.Get(ctx, key, receiver)
}

func (h *BaseHandler) DeleteLocalCache(ctx context.Context, key string) error {
	return h.client.Delete(ctx, key)
}

func (h *BaseHandler) DeleteRedisCache(ctx context.Context, key string) error {
	return h.client.Delete(ctx, key)
}
