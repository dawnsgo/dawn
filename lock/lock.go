package lock

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dawnsgo/dawn/log"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// LockMaker 获取分布式锁制造器
	LockMaker() Maker
}

var (
	// globalMaker 使用 atomic.Value 存储，无锁读取
	globalMaker atomic.Value // Maker

	// contextProvider 使用 atomic.Value 存储，无锁读取
	contextProvider atomic.Value // ContextProvider

	// writeMu 保护写入操作的原子性
	writeMu sync.Mutex
)

type Maker interface {
	// Make 制造一个Locker
	Make(name string) Locker
	// Close 关闭构建器
	Close() error
}

type Option struct {
	Once       bool          // 是否仅获取一次；默认阻塞地获取，直到获取成功
	Expiration time.Duration //
}

type Locker interface {
	// Acquire 获取锁
	Acquire(ctx context.Context) error
	// TryAcquire 尝试获取锁
	TryAcquire(ctx context.Context, expiration ...time.Duration) error
	// Release 释放锁
	Release(ctx context.Context) error
}

// contextProviderWrapper 包装器，解决 atomic.Value 存储 nil 接口的问题
type contextProviderWrapper struct {
	provider ContextProvider
}

// SetContextProvider 设置 Context 提供者（由 dawn 包调用）
func SetContextProvider(provider ContextProvider) {
	if provider == nil {
		contextProvider.Store((*contextProviderWrapper)(nil))
	} else {
		contextProvider.Store(&contextProviderWrapper{provider})
	}
}

// GetContextProvider 获取 Context 提供者
func GetContextProvider() ContextProvider {
	if v := contextProvider.Load(); v != nil {
		if w, ok := v.(*contextProviderWrapper); ok && w != nil {
			return w.provider
		}
	}
	return nil
}

// SetMaker 设置Locker制造商
func SetMaker(maker Maker) {
	if maker == nil {
		log.Warn("cannot set a nil lock-maker")
		return
	}

	writeMu.Lock()
	defer writeMu.Unlock()

	// 关闭旧的 maker
	if old := getGlobalMaker(); old != nil {
		if err := old.Close(); err != nil {
			log.Error("close lock-maker failed: %v", err)
		}
	}

	globalMaker.Store(maker)
}

// getGlobalMaker 获取全局 maker（无锁）
func getGlobalMaker() Maker {
	if v := globalMaker.Load(); v != nil {
		return v.(Maker)
	}
	return nil
}

// GetMaker 获取Locker制造商
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetMaker() Maker {
	if provider := GetContextProvider(); provider != nil {
		if maker := provider.LockMaker(); maker != nil {
			return maker
		}
	}
	return getGlobalMaker()
}

// getMaker 内部获取制造商（供其他函数调用）
func getMaker() Maker {
	return GetMaker()
}

// Make 制造一个Locker
func Make(name string) Locker {
	if maker := getMaker(); maker != nil {
		return maker.Make(name)
	}
	return nil
}

// Close 关闭构建器
func Close() error {
	writeMu.Lock()
	defer writeMu.Unlock()

	if maker := getGlobalMaker(); maker != nil {
		return maker.Close()
	}
	return nil
}
