// Package cache @Author  wangjian    2023/6/21 4:51 PM
package cache

import (
	"context"
	cacheConfig "github.com/JianWangEx/commonService/cache/config"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var (
	client *redisManager
	once   sync.Once
)

func GetRedisManager() Client {
	return client
}

func getRedisConn() (redis.UniversalClient, error) {
	// 获取redis config
	config := cacheConfig.GetRedisConfig()
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

func Init() (initErr error) {
	once.Do(func() {
		redisClient, err := getRedisConn()
		initErr = err
		client = &redisManager{
			redisClient: redisClient,
		}
	})
	return
}
