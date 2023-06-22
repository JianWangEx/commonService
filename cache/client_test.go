// Package cache @Author  wangjian    2023/6/21 11:45 PM
package cache

import (
	"context"
	"github.com/JianWangEx/commonService/cache/config"
	"testing"
)

func TestCache(t *testing.T) {
	path := "./config/redis_config.toml"
	err := config.InitRedisTomlConfig(path)
	if err != nil {
		panic(err)
	}
	err = Init()
	if err != nil {
		panic(err)
	}
	manager := GetRedisManager()
	err = manager.Add(context.TODO(), "test_redis_key", "123", 300000000000)
	if err != nil {
		panic(err)
	}
}
