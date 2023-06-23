// Package cache @Author  wangjian    2023/6/23 11:27 AM
package cache

import (
	"context"
	"github.com/JianWangEx/commonService/constant"
	logger "github.com/JianWangEx/commonService/log"
	"github.com/JianWangEx/commonService/util"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type addCacheLockParam struct {
	cacheKey       string
	cacheValueFunc func() string // 默认使用uuid
	timeout        time.Duration

	deleteAfterDone      bool
	blockingTimeout      time.Duration
	noLockRaiseException bool
	noLockReturn         constant.ErrorDict
}

type SetCacheLockParam func(param *addCacheLockParam)

func NewAddCacheLockParam(cacheKey string, opts ...SetCacheLockParam) *addCacheLockParam {
	// 默认的cacheLockParam
	param := &addCacheLockParam{
		cacheKey:             cacheKey,
		cacheValueFunc:       defaultCacheValueFunc,
		timeout:              defaultCacheLockTimeout,
		deleteAfterDone:      true,
		blockingTimeout:      defaultLockingTimeout,
		noLockRaiseException: false,
		noLockReturn:         constant.ErrorCodeErrorLock,
	}

	// 更新param
	for _, f := range opts {
		f(param)
	}

	return param
}

func CacheLockWithTimeout(timeout time.Duration) SetCacheLockParam {
	return func(param *addCacheLockParam) {
		param.timeout = timeout
	}
}

func CacheLockWithDeleteAfterDone(b bool) SetCacheLockParam {
	return func(param *addCacheLockParam) {
		param.deleteAfterDone = b
	}
}

func CacheLockWithBlockingTimeout(t time.Duration) SetCacheLockParam {
	return func(param *addCacheLockParam) {
		param.blockingTimeout = t
	}
}

func CacheLockWithNoLockRaiseException(b bool) SetCacheLockParam {
	return func(param *addCacheLockParam) {
		param.noLockRaiseException = b
	}
}

func CacheLockWithNoLockReturn(e constant.ErrorDict) SetCacheLockParam {
	return func(param *addCacheLockParam) {
		param.noLockReturn = e
	}
}

func CacheLockWithCacheValueFunc(f func() string) SetCacheLockParam {
	return func(param *addCacheLockParam) {
		param.cacheValueFunc = f
	}
}

type AddCacheLockOperator func(ctx context.Context) (i interface{}, e error)

// AddCacheLockHandler
//
//	@Description:对o function进行修饰，根据params添加缓存锁，并返回o执行结果
//	@param ctx
//	@param o
//	@param params
//	@return interface{}
//	@return error
func AddCacheLockHandler(ctx context.Context, o AddCacheLockOperator, params ...*addCacheLockParam) (interface{}, error) {
	if existLockContainer := ctx.Value(addCacheLockCtxKey); existLockContainer == nil {
		m := new(sync.Map)
		ctx = context.WithValue(ctx, addCacheLockCtxKey, m)
	}
	fn := addCacheLockHandler(o, params...)
	return fn(ctx)
}

func addCacheLockHandler(o AddCacheLockOperator, params ...*addCacheLockParam) AddCacheLockOperator {
	for i := len(params) - 1; i >= 0; i-- {
		o = addCacheLock(o, params[i])
	}
	return o
}

func addCacheLock(o AddCacheLockOperator, param *addCacheLockParam) AddCacheLockOperator {
	return func(ctx context.Context) (i interface{}, e error) {
		log := logger.CtxSugar(ctx)
		key := param.cacheKey
		value := param.cacheValueFunc()
		lock := newLock(key, param, value)
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("addCacheLock panic,cache_key=%+v, err=%+v", key, err)
				e = errors.WithStack(constant.CommonErrorServer.WithMsgF("add_cache_lock_panic,err=%+v", err))
			}
			lock.release(ctx)
		}()

		if lockSuccess := lock.acquire(ctx); !lockSuccess {
			log.Errorf("func blocked by cache lock|cacheKey=%s, cacheValue=%+v", key, value)
			if param.noLockRaiseException {
				log.Errorf("func blocked by cache lock and param.noLockRaiseException is true|return and do nothing|cacheKey=%s, cacheValue=%+v", key, value)
				return nil, errors.WithStack(constant.CommonErrorServer.WithMsg("blocked by cache lock"))
			}
			return param.noLockReturn, errors.WithStack(constant.CommonErrorServer.WithMsg("blocked by cache lock"))
		}
		i, e = o(ctx)
		return i, e
	}
}

type cacheLock struct {
	cacheKey        string
	cacheValue      string
	cacheTimeout    time.Duration
	deleteAfterDone bool
	lockSuccess     bool
	blockingTimeout time.Duration
	didInheritLock  bool
}

func newLock(key string, param *addCacheLockParam, value string) *cacheLock {
	return &cacheLock{
		cacheKey:        key,
		cacheValue:      value,
		cacheTimeout:    param.timeout,
		deleteAfterDone: param.deleteAfterDone,
		blockingTimeout: param.blockingTimeout,
		// 新锁没有继承标记
		didInheritLock: false,
	}
}

// acquire 获得锁
func (l *cacheLock) acquire(ctx context.Context) bool {
	log := logger.CtxSugar(ctx)
	enterTime := time.Now()
	if l.inheritLockFromCurrentContext(ctx) {
		_ = time.Since(enterTime).Milliseconds()
		return true
	}

	numOfRetry := 0
	c := GetCacheManager()
	for {
		startTime := time.Now()
		err := c.Add(ctx, l.cacheKey, l.cacheValue, l.cacheTimeout)
		timeToken := time.Since(startTime).Milliseconds()
		if err == nil {
			l.lockSuccess = true
		} else {
			log.Errorf("cache lock acquire call error|cache_key=%s,cache_value=%s,err=%v", l.cacheKey, l.cacheValue, err)
			l.lockSuccess = false
		}
		if timeToken > warningThreshold {
			log.Warnf("cache_is_too_slow|cache_key=%s,cache_value=%s,time_token=%d[ms]", l.cacheKey, l.cacheValue, timeToken)
		}

		if l.lockSuccess {
			// set lock to context
			addCacheLockContext := ctx.Value(addCacheLockCtxKey)
			if addCacheLockContext != nil {
				if m, ok := addCacheLockContext.(*sync.Map); ok {
					m.Store(l.cacheKey, l.cacheValue)
				}
			}
			return l.lockSuccess
		} else {
			timeToken := time.Since(enterTime)
			if timeToken <= l.blockingTimeout {
				waitTime := util.MinDuration(l.blockingTimeout-timeToken, (1<<numOfRetry)*time.Second/lockPollingIntervalDiv)
				time.Sleep(waitTime)
				numOfRetry++
				continue
			}
			nowCacheValue := l.getNowCacheValue(ctx, c)
			log.Errorf("acquire lock failed|cacheKey=%s,cacheValue=%s,cacheTimeout=%d,nowCacheValue=%+v", l.cacheKey, l.cacheValue, l.cacheTimeout.Milliseconds(), nowCacheValue)
			return false
		}

	}
}

func (l *cacheLock) inheritLockFromCurrentContext(ctx context.Context) bool {
	// 从parent context(sync.Map)中载入lock
	addCacheLockContext := ctx.Value(addCacheLockCtxKey)
	ok, val := l.getValueFromCtx(addCacheLockContext)
	if !ok {
		return false
	}
	c := GetCacheManager()
	nowCacheValue := l.getNowCacheValue(ctx, c)
	if val == *nowCacheValue {
		l.didInheritLock = true
		l.cacheValue = val
		return true
	}
	return false

}

func (l *cacheLock) getValueFromCtx(addCacheLockContext interface{}) (bool, string) {
	if addCacheLockContext == nil {
		return false, ""
	}
	if m, ok := addCacheLockContext.(*sync.Map); !ok {
		return false, ""
	} else if contextCacheValue, ok := m.Load(l.cacheKey); ok {
		if value, ok := contextCacheValue.(string); ok {
			return true, value
		}
	}
	return false, ""
}

func (l *cacheLock) getNowCacheValue(ctx context.Context, c Client) *string {
	nowCacheValue := new(string)
	err := c.Get(ctx, l.cacheKey, nowCacheValue)
	if err != nil {
		nowCacheValue = util.StringToPtr("?")
	}
	return nowCacheValue
}

func (l *cacheLock) release(ctx context.Context) {
	log := logger.CtxSugar(ctx)
	c := GetCacheManager()
	if l.lockSuccess {
		// 如果使用map删除它
		addCacheLockContext := ctx.Value(addCacheLockCtxKey)
		if addCacheLockContext != nil {
			if m, ok := addCacheLockContext.(*sync.Map); ok {
				m.Delete(l.cacheKey)
			}
		}

		// 如果设置了完成后删除
		if l.deleteAfterDone {
			nowCacheValue := l.getNowCacheValue(ctx, c)
			if *nowCacheValue == l.cacheValue {
				err := c.Delete(ctx, l.cacheKey)
				if err != nil {
					log.Errorf("cache lock, release delete error|cacheKey=%s, cacheValue=%+v, cacheTimeout=%d[ms]", l.cacheKey, l.cacheValue, l.cacheTimeout.Milliseconds())
				}
			} else {
				log.Errorf("cache lock, release|cache lock already expire|cacheKey=%s, cacheValue=%+v, cacheTimeout=%d[ms]", l.cacheKey, l.cacheValue, l.cacheTimeout.Milliseconds())
			}
			l.lockSuccess = false
		}
	} else if l.didInheritLock {
		return
	}
}
