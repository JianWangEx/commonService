// Package cache @Author  wangjian    2023/6/21 4:51 PM
package cache

import (
	"context"
	cacheConfig "github.com/JianWangEx/commonService/cache/config"
	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	client *cacheManager
	once   sync.Once
)

type cacheManager struct {
	redisClient      redis.UniversalClient
	localCacheClient *LocalCacheManager
}

func GetCacheManager() Client {
	return client
}

func getRedisConn() (redis.UniversalClient, error) {
	// 获取redis config
	config := cacheConfig.GetCacheConfig()
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:                 config.Addrs,
		ClientName:            config.ClientName,
		DB:                    config.DB,
		Dialer:                config.Dialer,
		OnConnect:             config.OnConnect,
		Protocol:              config.Protocol,
		Username:              config.Username,
		Password:              config.Password,
		SentinelUsername:      config.SentinelUsername,
		SentinelPassword:      config.SentinelPassword,
		MaxRetries:            config.MaxRetries,
		MinRetryBackoff:       config.MinRetryBackoff,
		MaxRetryBackoff:       config.MaxRetryBackoff,
		DialTimeout:           time.Duration(config.DialTimeout),
		ReadTimeout:           time.Duration(config.ReadTimeout),
		WriteTimeout:          time.Duration(config.WriteTimeout),
		ContextTimeoutEnabled: config.ContextTimeoutEnabled,
		PoolFIFO:              config.PoolFIFO,
		PoolSize:              config.PoolSize,
		PoolTimeout:           config.PoolTimeout,
		MinIdleConns:          config.MinIdleConns,
		MaxIdleConns:          config.MaxIdleConns,
		ConnMaxIdleTime:       config.ConnMaxIdleTime,
		ConnMaxLifetime:       config.ConnMaxLifetime,
		TLSConfig:             config.TLSConfig,
		MaxRedirects:          config.MaxRetries,
		ReadOnly:              config.ReadOnly,
		RouteByLatency:        config.RouteByLatency,
		RouteRandomly:         config.RouteRandomly,
		MasterName:            config.MasterName,
	})

	if err := redisClient.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}
	return redisClient, nil
}

func getLocalCache() *LocalCacheManager {
	// 获取local cache config
	config := cacheConfig.GetCacheConfig()
	return &LocalCacheManager{
		cache.New(config.DefaultExpiration, config.CleanupInterval),
	}
}

func Init() (initErr error) {
	once.Do(func() {
		redisClient, err := getRedisConn()
		initErr = err
		lc := getLocalCache()
		client = &cacheManager{
			redisClient:      redisClient,
			localCacheClient: lc,
		}
	})
	return
}
