package cache

import (
	"context"
	"sync"
	"time"

	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// Cache 获取缓存
	Cache() Cache
}

var (
	globalCache     Cache
	contextProvider ContextProvider
	providerMu      sync.RWMutex
)

type SetValueFunc func() (any, error)

type Cache interface {
	// Has 检测缓存是否存在
	Has(ctx context.Context, key string) (bool, error)
	// Get 获取缓存值
	Get(ctx context.Context, key string, def ...any) Result
	// Set 设置缓存值
	Set(ctx context.Context, key string, value any, expiration ...time.Duration) error
	// GetSet 获取设置缓存值
	GetSet(ctx context.Context, key string, fn SetValueFunc) Result
	// Delete 删除缓存
	Delete(ctx context.Context, keys ...string) (int64, error)
	// IncrInt 整数自增
	IncrInt(ctx context.Context, key string, value int64) (int64, error)
	// IncrFloat 浮点数自增
	IncrFloat(ctx context.Context, key string, value float64) (float64, error)
	// DecrInt 整数自减
	DecrInt(ctx context.Context, key string, value int64) (int64, error)
	// DecrFloat 浮点数自减
	DecrFloat(ctx context.Context, key string, value float64) (float64, error)
	// AddPrefix 添加Key前缀
	AddPrefix(key string) string
	// Client 获取客户端
	Client() any
	// Close 关闭缓存
	Close() error
}

// SetContextProvider 设置 Context 提供者（由 dawn 包调用）
func SetContextProvider(provider ContextProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	contextProvider = provider
}

// GetContextProvider 获取 Context 提供者
func GetContextProvider() ContextProvider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return contextProvider
}

// SetCache 设置缓存
func SetCache(cache Cache) {
	if cache == nil {
		log.Warn("cannot set a nil cache")
		return
	}

	providerMu.Lock()
	defer providerMu.Unlock()

	if globalCache != nil {
		if err := globalCache.Close(); err != nil {
			log.Error("close cache failed: %v", err)
		}
	}

	globalCache = cache
}

// GetCache 获取缓存
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetCache() Cache {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if ca := contextProvider.Cache(); ca != nil {
			return ca
		}
	}
	return globalCache
}

// getCache 内部获取缓存（供其他函数调用）
func getCache() Cache {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if ca := contextProvider.Cache(); ca != nil {
			return ca
		}
	}
	return globalCache
}

// Has 检测缓存是否存在
func Has(ctx context.Context, key string) (bool, error) {
	if ca := getCache(); ca != nil {
		return ca.Has(ctx, key)
	}
	return false, errors.ErrMissingCacheInstance
}

// Get 获取缓存值
func Get(ctx context.Context, key string, def ...any) Result {
	if ca := getCache(); ca != nil {
		return ca.Get(ctx, key, def...)
	}
	return NewResult(nil, errors.ErrMissingCacheInstance)
}

// Set 设置缓存值
func Set(ctx context.Context, key string, value any, expiration ...time.Duration) error {
	if ca := getCache(); ca != nil {
		return ca.Set(ctx, key, value, expiration...)
	}
	return errors.ErrMissingCacheInstance
}

// GetSet 获取设置缓存值
func GetSet(ctx context.Context, key string, fn SetValueFunc) Result {
	if ca := getCache(); ca != nil {
		return ca.GetSet(ctx, key, fn)
	}
	return NewResult(nil, errors.ErrMissingCacheInstance)
}

// Delete 删除缓存
func Delete(ctx context.Context, keys ...string) (int64, error) {
	if ca := getCache(); ca != nil {
		return ca.Delete(ctx, keys...)
	}
	return 0, errors.ErrMissingCacheInstance
}

// IncrInt 整数自增
func IncrInt(ctx context.Context, key string, value int64) (int64, error) {
	if ca := getCache(); ca != nil {
		return ca.IncrInt(ctx, key, value)
	}
	return 0, errors.ErrMissingCacheInstance
}

// IncrFloat 浮点数自增
func IncrFloat(ctx context.Context, key string, value float64) (float64, error) {
	if ca := getCache(); ca != nil {
		return ca.IncrFloat(ctx, key, value)
	}
	return 0, errors.ErrMissingCacheInstance
}

// DecrInt 整数自减
func DecrInt(ctx context.Context, key string, value int64) (int64, error) {
	if ca := getCache(); ca != nil {
		return ca.DecrInt(ctx, key, value)
	}
	return 0, errors.ErrMissingCacheInstance
}

// DecrFloat 浮点数自减
func DecrFloat(ctx context.Context, key string, value float64) (float64, error) {
	if ca := getCache(); ca != nil {
		return ca.DecrFloat(ctx, key, value)
	}
	return 0, errors.ErrMissingCacheInstance
}

// AddPrefix 添加Key前缀
func AddPrefix(key string) string {
	if ca := getCache(); ca != nil {
		return ca.AddPrefix(key)
	}
	return ""
}

// Client 获取客户端
func Client() any {
	if ca := getCache(); ca != nil {
		return ca.Client()
	}
	return nil
}

// Close 关闭缓存
func Close() error {
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalCache != nil {
		return globalCache.Close()
	}
	return nil
}
