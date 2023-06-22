// Package cache @Author  wangjian    2023/6/21 11:45 PM
package cache

import (
	"context"
	"fmt"
	"github.com/JianWangEx/commonService/cache/config"
	"github.com/JianWangEx/commonService/util"
	"testing"
)

func TestCache(t *testing.T) {
	path := "./config/redis_config.toml"
	err := config.InitCacheTomlConfig(path)
	if err != nil {
		panic(err)
	}
	err = Init()
	if err != nil {
		panic(err)
	}
	manager := GetCacheManager()
	err = manager.Add(context.TODO(), "test_redis_key.main", "123", 300000000000)
	if err != nil {
		panic(err)
	}
	val := util.StringToPtr("")
	err = manager.Get(context.TODO(), "test_redis_key.main", val)
	if err != nil {
		panic(err)
	}
	fmt.Println(val)
}
