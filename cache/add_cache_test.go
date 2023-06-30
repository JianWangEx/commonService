// Package cache @Author  wangjian    2023/6/30 9:00 AM
package cache

import (
	"context"
	"fmt"
	"testing"
)

func TestAddCache(t *testing.T) {
	//path := "./config/redis_config.toml"
	//err := config.InitCacheTomlConfig(path)
	//if err != nil {
	//	panic(err)
	//}
	//err = Init()
	//if err != nil {
	//	panic(err)
	//}
	//
	//key := uuid.NewString()
	//fmt.Println(key)
	//d1 := NewAddCacheParam(
	//	key,
	//	WithCacheStore(Local),
	//	WithEncodeKeyType(Md5),
	//)
	//d2 := NewAddCacheParam(
	//	key,
	//	WithCacheStore(Main),
	//	WithEncodeKeyType(Md5),
	//)
	//result := new(map[string]string)
	//err = AddCacheHandle(context.TODO(), result, testFunc, d1, d2)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("success")
	//fmt.Println(result)
}

func testFunc(ctx context.Context, testValue interface{}) (interface{}, error) {
	m := map[string]string{
		"name": "cat",
		"age":  "1",
	}
	fmt.Println(m)
	return m, nil
}
