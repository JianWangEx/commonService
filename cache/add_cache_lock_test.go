// Package cache @Author  wangjian    2023/6/23 2:44 PM
package cache_test

import (
	"fmt"
	"testing"
)

func mainTestLockKey() string {
	return fmt.Sprintf("mainTestLockKey")
}

func TestAddCacheLockDecorator(t *testing.T) {
	//path := "./config/redis_config.toml"
	//err := config.InitCacheTomlConfig(path)
	//if err != nil {
	//	panic(err)
	//}
	//err = cache.Init()
	//if err != nil {
	//	panic(err)
	//}
	//
	//d1 := cache.NewAddCacheLockParam(
	//	mainTestLockKey(),
	//	cache.CacheLockWithTimeout(5*time.Minute),
	//	cache.CacheLockWithBlockingTimeout(5*time.Second),
	//	cache.CacheLockWithDeleteAfterDone(true),
	//	cache.CacheLockWithNoLockRaiseException(false),
	//	cache.CacheLockWithNoLockReturn(constant.ErrorCodeErrorLock),
	//)
	//i, err := cache.AddCacheLockHandler(context.TODO(), func(ctx context.Context) (i interface{}, e error) {
	//	fmt.Println("exec")
	//	return i, e
	//}, d1)
	//
	//fmt.Print(i, err)
}
