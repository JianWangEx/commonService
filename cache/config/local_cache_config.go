// Package config @Author  wangjian    2023/6/22 6:12 PM
package config

import (
	"time"
)

type LocalCacheConfig struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}
