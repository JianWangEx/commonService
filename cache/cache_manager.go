// Package cache @Author  wangjian    2023/6/22 10:17 PM
package cache

import (
	"context"
	"encoding/json"
	"github.com/JianWangEx/commonService/constant"
	"github.com/JianWangEx/commonService/util"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"reflect"
	"strings"
	"time"
)

func (c *cacheManager) Get(ctx context.Context, key string, receiver interface{}) error {
	// 校验receiver类型
	rv := reflect.ValueOf(receiver)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return constant.ErrorNilReceiverOrNotPtr
	}

	storage := getStorage(key)

	switch storage {
	case Local:
		val, found := c.localCacheClient.Get(ctx, key)
		if !found {
			return constant.ErrorCacheMiss
		}
		return util.SetValue(val, receiver)
	default: // default is main
		result := c.redisClient.Get(ctx, key)
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
		// only json code type
		return json.Unmarshal(data, receiver)
	}
}

func (c *cacheManager) Set(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	storage := getStorage(key)
	switch storage {
	case Local:
		c.localCacheClient.Set(ctx, key, util.GetValue(value), expired)
		return nil
	default: // default is main
		data, err := json.Marshal(value)
		if err != nil {
			return errors.Wrap(err, "json marshal error")
		}
		result := c.redisClient.Set(ctx, key, data, expired)
		if err := result.Err(); err != nil {
			return errors.Wrap(err, "redis cache error")
		}
		return nil
	}
}

func (c *cacheManager) Add(ctx context.Context, key string, value interface{}, expired time.Duration) error {
	storage := getStorage(key)
	switch storage {
	case Local:
		if err := c.localCacheClient.Add(ctx, key, util.GetValue(value), expired); err != nil {
			return constant.ErrorFailedOperation
		}
	default: // default is main
		data, err := json.Marshal(value)
		if err != nil {
			return errors.Wrap(err, "json marshal error")
		}
		result := c.redisClient.SetNX(ctx, key, data, expired)
		if err := result.Err(); err != nil {
			return errors.Wrap(err, "redis cache error")
		}
		if !result.Val() {
			return constant.ErrorFailedOperation
		}
	}
	return nil
}

func getStorage(key string) Storage {
	storage := Main
	splits := strings.Split(key, ".")
	suffix := splits[len(splits)-1]
	if suffix == Local.name() {
		storage = Local
	}
	return storage
}
