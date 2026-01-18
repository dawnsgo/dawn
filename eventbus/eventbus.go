package eventbus

import (
	"context"
	"sync"

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
	globalEventbus  Eventbus
	contextProvider ContextProvider
	providerMu      sync.RWMutex
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

// SetEventbus 设置事件总线
func SetEventbus(eb Eventbus) {
	if eb == nil {
		log.Warn("cannot set a nil eventbus")
		return
	}

	providerMu.Lock()
	defer providerMu.Unlock()

	if globalEventbus != nil {
		if err := globalEventbus.Close(); err != nil {
			log.Errorf("the old eventbus close failed: %v", err)
		}
	}

	globalEventbus = eb
}

// GetEventbus 获取事件总线
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetEventbus() Eventbus {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if eb := contextProvider.Eventbus(); eb != nil {
			return eb
		}
	}
	return globalEventbus
}

// getEventbus 内部获取事件总线（供其他函数调用）
func getEventbus() Eventbus {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if eb := contextProvider.Eventbus(); eb != nil {
			return eb
		}
	}
	return globalEventbus
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
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalEventbus != nil {
		return globalEventbus.Close()
	}
	return nil
}
