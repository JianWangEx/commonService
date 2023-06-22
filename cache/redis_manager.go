// Package cache @Author  wangjian    2023/6/21 7:58 PM
package cache

import (
	"context"
	"encoding/json"
	"github.com/JianWangEx/commonService/constant"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"reflect"
	"time"
)

type redisManager struct {
	redisClient redis.UniversalClient
}

func (r *redisManager) Get(ctx context.Context, key string, receiver interface{}) error {
	rv := reflect.ValueOf(receiver)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return constant.ErrorNilReceiverOrNotPtr
	}
	result := r.redisClient.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return constant.ErrorCacheMiss
		}
		return errors.Wrap(err, "redis cache error")
	}

	if result.Val() == "null" {
		return constant.ErrorCacheMiss
	}

	data, err := result.Bytes()
	if err != nil {
		return errors.Wrap(err, "redis cache error")
	}
	return json.Unmarshal(data, receiver)
}

func (r *redisManager) Set(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}
	result := r.redisClient.Set(ctx, key, data, expired)
	if err := result.Err(); err != nil {
		return errors.Wrap(err, "redis cache error")
	}
	return nil
}

func (r *redisManager) Add(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "json marshal error")
	}
	result := r.redisClient.SetNX(ctx, key, data, expired)
	if err := result.Err(); err != nil {
		return errors.Wrap(err, "redis cache error")
	}
	if !result.Val() {
		return constant.ErrorFailedOperation
	}
	return nil
}
