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
	defaultCacheTimeout       = 30
	defaultCacheTimeoutSecond = defaultCacheTimeout * time.Second
	cacheBust                 = false

	maxLogValueLength = 1000
)

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

type EncodeKeyType string

const (
	Utf8   EncodeKeyType = "utf-8"
	Md5    EncodeKeyType = "md5"
	Base64 EncodeKeyType = "base64"
)
