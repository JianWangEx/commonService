// Package cache @Author  wangjian    2023/6/29 5:26 PM
package cache

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/JianWangEx/commonService/constant"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/JianWangEx/commonService/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

type CanStoreResultFunc func(v interface{}) bool

type AddCacheParam struct {
	cacheKey       string
	timeout        time.Duration
	cacheBust      bool // 确保获取最新的数据通过判断cache bust 破坏缓存
	encodeKeyType  EncodeKeyType
	cacheStore     Storage
	canStoreResult CanStoreResultFunc
}

func NewAddCacheParam(cacheKey string, options ...SetParam) *AddCacheParam {
	// set default params
	var defaultCanStoreResult CanStoreResultFunc = func(v interface{}) bool {
		if util.IsNil(v) {
			return false
		}
		return true
	}

	params := &AddCacheParam{
		cacheKey:       cacheKey,
		timeout:        defaultCacheTimeoutSecond,
		cacheBust:      cacheBust,
		encodeKeyType:  Utf8,
		cacheStore:     Main,
		canStoreResult: defaultCanStoreResult,
	}

	// set options
	for _, paramSet := range options {
		paramSet(params)
	}
	return params
}

type SetParam func(param *AddCacheParam)

func WithTimeout(timeout time.Duration) SetParam {
	return func(params *AddCacheParam) {
		params.timeout = timeout
	}
}

func WithCacheBust(cacheBust bool) SetParam {
	return func(param *AddCacheParam) {
		param.cacheBust = cacheBust
	}
}

func WithEncodeKeyType(encodeKeyType EncodeKeyType) SetParam {
	return func(param *AddCacheParam) {
		param.encodeKeyType = encodeKeyType
	}
}

func WithCacheStore(cacheStore Storage) SetParam {
	return func(param *AddCacheParam) {
		param.cacheStore = cacheStore
	}
}

func WithCanStoreResult(canStoreResult CanStoreResultFunc) SetParam {
	return func(param *AddCacheParam) {
		param.canStoreResult = canStoreResult
	}
}

// AddCacheOperator
// receiver must be a ptr and not nil, like new(string)
// your func return value should set to receiver
type AddCacheOperator func(ctx context.Context, receiver interface{}) (i interface{}, e error)

func AddCacheHandle(ctx context.Context, receiver interface{}, f AddCacheOperator, decoratorParams ...*AddCacheParam) error {
	rv := reflect.ValueOf(receiver)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return constant.ErrorNilReceiverOrNotPtr
	}
	fn := addCacheHandler(f, decoratorParams...)
	_, err := fn(ctx, receiver)
	return err
}

func addCacheHandler(m AddCacheOperator, params ...*AddCacheParam) AddCacheOperator {
	for i := len(params) - 1; i >= 0; i-- {
		m = addCache(m, params[i])
	}
	return m
}

func addCache(m AddCacheOperator, param *AddCacheParam) AddCacheOperator {
	return func(ctx context.Context, receiver interface{}) (i interface{}, e error) {
		log := logger.CtxSugar(ctx)
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("addcCache panic,param=%+v, err=%+v", param, err)
				e = errors.WithStack(constant.CommonErrorServer.WithMsgF("add cache panic,err=%+v", err))
			}
		}()
		key := param.cacheKey
		keyRaw := key
		key = getCacheKey(param, key)
		// 尝试从cache中获取值
		cacheFound := tryFindCache(ctx, receiver, param, key, keyRaw)
		c := GetCacheManager()
		if cacheFound {
			return receiver, nil
		} else {
			// get from real func
			i, e = m(ctx, receiver)
			if e != nil {
				if e == constant.ErrorAddCacheGotNilResult {
					log.Warnf("addCache call real function, key=%v,result is nil", keyRaw)
				} else {
					log.Warnf("addCache call func m(ctx, receiver), key=%v,err=%+v", keyRaw, e)
				}
				return i, e
			}

			// i == nil
			if util.IsNil(i) {
				log.Warnf("addCache call real function, key=%v,result is nil", keyRaw)
				return i, constant.ErrorAddCacheGotNilResult
			}
			e = util.DeepCopy(receiver, i)

			if !param.canStoreResult(i) {
				return i, e
			}
			setCache(ctx, i, c, key, param, log, keyRaw)
		}
		return i, e
	}
}

func tryFindCache(ctx context.Context, receiver interface{}, param *AddCacheParam, key string, keyRaw string) bool {
	log := logger.CtxSugar(ctx)
	c := GetCacheManager()
	cacheFound := false

	if param.cacheBust {
		err := c.Delete(ctx, key)
		if err != nil {
			log.Warnf("addCache Delete failed|key=%+v, err=%+v", key, err)
		}
	} else {
		readStartTime := time.Now()
		err := c.Get(ctx, key, receiver)
		readElapsedTime := time.Since(readStartTime).Milliseconds()
		if err != nil {
			log.Warnf("addCache get from cache failed|key=%+v, err=%+v", key, err)
		} else {
			cacheFound = true
		}
		logForReadOvertimeCost(readElapsedTime, receiver, log, keyRaw, key)
	}
	return cacheFound
}

func setCache(ctx context.Context, value interface{}, c Client, key string, param *AddCacheParam, log *zap.SugaredLogger, keyRaw string) {
	writeStartTime := time.Now()
	err := c.Set(ctx, key, value, param.timeout)
	writeElapsedTime := time.Since(writeStartTime).Milliseconds()
	if err != nil {
		log.Warnf("addCache Set to cache failed|key=%+v,keyRaw=%+v, err=%+v", key, keyRaw, err)
	}
	logForWriteOvertimeCost(writeElapsedTime, value, log, keyRaw, key)
}

func logForWriteOvertimeCost(costTime int64, receiver interface{}, log *zap.SugaredLogger, keyRaw, key string) {
	if costTime > warningThreshold {
		logValue := getValueForLog(receiver)
		log.Warnf("cache write is too slow|key_raw=%s,key=%s,value=%s,time=%d", keyRaw, key, logValue, costTime)
	}
}

func logForReadOvertimeCost(costTime int64, receiver interface{}, log *zap.SugaredLogger, keyRaw string, key string) {
	if costTime > warningThreshold {
		logValue := getValueForLog(receiver)
		log.Warnf("cache_read_is_too_slow|key_raw=%s,key=%s,value=%s,time=%d", keyRaw, key, logValue, costTime)
	}
}

func getValueForLog(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "GetValueError"
	}
	rawStr := string(bytes)
	if len(rawStr) > maxLogValueLength {
		rawStr = strings.Join([]string{rawStr[:maxLogValueLength], "...(truncated)"}, "")
	}
	return rawStr
}

func getCacheKey(param *AddCacheParam, key string) string {
	switch param.encodeKeyType {
	case Md5:
		key = md5Key(key)
	case Base64:
		key = base64.StdEncoding.EncodeToString([]byte(key))
	}
	return key
}

func md5Key(key string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(key))
	cipherRecordByte := md5Ctx.Sum(nil)
	cipherRecordStr := hex.EncodeToString(cipherRecordByte)
	return cipherRecordStr
}
