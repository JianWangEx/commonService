// Package cache @Author  wangjian    2023/6/22 10:27 PM
package cache

import (
	"github.com/google/uuid"
	"time"
)

type Storage string

const (
	Main  Storage = "main"
	Local Storage = "local"
)

func (s Storage) name() string {
	return string(s)
}

const (
	defaultCacheLockTimeout = 300 * time.Second
	defaultLockingTimeout   = 0

	warningThreshold = 1000 // unit milliseconds

	// 锁轮训间隔划分
	lockPollingIntervalDiv = 100

	addCacheLockCtxKey = "addCacheLockCtxKey"
)

var defaultCacheValueFunc = func() string {
	return uuid.NewString()
}
