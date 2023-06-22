// Package config @Author  wangjian    2023/6/22 11:50 PM
package config

import (
	"github.com/JianWangEx/commonService/util"
	"sync"
)

var (
	tomlOnce sync.Once

	cacheConfigObj = &CacheConfig{}
)

type CacheConfig struct {
	RedisConfig
	LocalCacheConfig
}

func InitCacheTomlConfig(path string) (err error) {
	tomlOnce.Do(func() {
		err = util.ParseTomlConfig(path, cacheConfigObj)
	})
	return err
}

func GetCacheConfig() *CacheConfig {
	return cacheConfigObj
}
