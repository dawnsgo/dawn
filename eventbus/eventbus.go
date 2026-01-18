package eventbus

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/eventbus/internal"
	"github.com/dawnsgo/dawn/log"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// Eventbus 获取事件总线
	Eventbus() Eventbus
}

var (
	// globalEventbus 使用 atomic.Value 存储，无锁读取
	globalEventbus atomic.Value // Eventbus

	// contextProvider 使用 atomic.Value 存储，无锁读取
	contextProvider atomic.Value // ContextProvider

	// writeMu 保护写入操作的原子性
	writeMu sync.Mutex
)

type (
	Event        = internal.Event
	EventHandler = internal.EventHandler
)

type Eventbus interface {
	// Close 关闭事件总线
	Close() error
	// Publish 发布事件
	Publish(ctx context.Context, topic string, message any) error
	// Subscribe 订阅事件
	Subscribe(ctx context.Context, topic string, handler EventHandler) error
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, topic string, handler EventHandler) error
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

// SetEventbus 设置事件总线
func SetEventbus(eb Eventbus) {
	if eb == nil {
		log.Warn("cannot set a nil eventbus")
		return
	}

	writeMu.Lock()
	defer writeMu.Unlock()

	// 关闭旧的 eventbus
	if old := getGlobalEventbus(); old != nil {
		if err := old.Close(); err != nil {
			log.Errorf("the old eventbus close failed: %v", err)
		}
	}

	globalEventbus.Store(eb)
}

// getGlobalEventbus 获取全局事件总线（无锁）
func getGlobalEventbus() Eventbus {
	if v := globalEventbus.Load(); v != nil {
		return v.(Eventbus)
	}
	return nil
}

// GetEventbus 获取事件总线
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetEventbus() Eventbus {
	if provider := GetContextProvider(); provider != nil {
		if eb := provider.Eventbus(); eb != nil {
			return eb
		}
	}
	return getGlobalEventbus()
}

// getEventbus 内部获取事件总线（供其他函数调用）
func getEventbus() Eventbus {
	return GetEventbus()
}

// Publish 发布事件
func Publish(ctx context.Context, topic string, message any) error {
	if eb := getEventbus(); eb != nil {
		return eb.Publish(ctx, topic, message)
	}
	return errors.ErrMissingEventbusInstance
}

// Subscribe 订阅事件
func Subscribe(ctx context.Context, topic string, handler EventHandler) error {
	if eb := getEventbus(); eb != nil {
		return eb.Subscribe(ctx, topic, handler)
	}
	return errors.ErrMissingEventbusInstance
}

// Unsubscribe 取消订阅
func Unsubscribe(ctx context.Context, topic string, handler EventHandler) error {
	if eb := getEventbus(); eb != nil {
		return eb.Unsubscribe(ctx, topic, handler)
	}
	return errors.ErrMissingEventbusInstance
}

// Close 关闭事件总线
func Close() error {
	writeMu.Lock()
	defer writeMu.Unlock()

	if eb := getGlobalEventbus(); eb != nil {
		return eb.Close()
	}
	return nil
}
