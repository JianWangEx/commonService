// Package config @Author  wangjian    2023/6/22 6:12 PM
package config

type LocalCacheConfig struct {
	DefaultExpiration int // time.Minute
	CleanupInterval   int // time.Minute
}
