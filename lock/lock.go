package lock

import (
	"context"
	"sync"
	"time"

	"github.com/dawnsgo/dawn/log"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// LockMaker 获取分布式锁制造器
	LockMaker() Maker
}

var (
	globalMaker     Maker
	contextProvider ContextProvider
	providerMu      sync.RWMutex
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

// SetMaker 设置Locker制造商
func SetMaker(maker Maker) {
	if maker == nil {
		log.Warn("cannot set a nil lock-maker")
		return
	}

	providerMu.Lock()
	defer providerMu.Unlock()

	if globalMaker != nil {
		if err := globalMaker.Close(); err != nil {
			log.Error("close lock-maker failed: %v", err)
		}
	}

	globalMaker = maker
}

// GetMaker 获取Locker制造商
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetMaker() Maker {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if maker := contextProvider.LockMaker(); maker != nil {
			return maker
		}
	}
	return globalMaker
}

// getMaker 内部获取制造商（供其他函数调用）
func getMaker() Maker {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if maker := contextProvider.LockMaker(); maker != nil {
			return maker
		}
	}
	return globalMaker
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
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalMaker != nil {
		return globalMaker.Close()
	}
	return nil
}
