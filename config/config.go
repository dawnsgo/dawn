// Package config 提供统一的配置管理接口和实现。
// 支持从多种配置源（文件、etcd、consul、nacos）加载和监听配置项。
package config

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/dawnsgo/dawn/core/value"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// Configurator 获取配置器
	Configurator() Configurator
}

var (
	// globalConfigurator 使用 atomic.Value 存储，无锁读取
	globalConfigurator atomic.Value // Configurator

	// contextProvider 使用 atomic.Value 存储，无锁读取
	contextProvider atomic.Value // ContextProvider

	// writeMu 保护写入操作的原子性
	writeMu sync.Mutex
)

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

// SetConfigurator 设置配置器
func SetConfigurator(configurator Configurator) {
	writeMu.Lock()
	defer writeMu.Unlock()

	// 关闭旧的 configurator
	if old := getGlobalConfigurator(); old != nil {
		old.Close()
	}
	if configurator != nil {
		globalConfigurator.Store(configurator)
	}
}

// getGlobalConfigurator 获取全局配置器（无锁）
func getGlobalConfigurator() Configurator {
	if v := globalConfigurator.Load(); v != nil {
		return v.(Configurator)
	}
	return nil
}

// GetConfigurator 获取配置器
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetConfigurator() Configurator {
	if provider := GetContextProvider(); provider != nil {
		if configurator := provider.Configurator(); configurator != nil {
			return configurator
		}
	}
	return getGlobalConfigurator()
}

// getConfigurator 内部获取配置器（供其他函数调用）
func getConfigurator() Configurator {
	return GetConfigurator()
}

// SetConfiguratorWithSources 通过设置配置源来设置配置器
func SetConfiguratorWithSources(sources ...Source) {
	SetConfigurator(NewConfigurator(WithSources(sources...)))
}

// Has 检测多个匹配规则中是否存在配置
func Has(pattern string) bool {
	if configurator := getConfigurator(); configurator != nil {
		return configurator.Has(pattern)
	}
	return false
}

// Get 获取配置值
func Get(pattern string, def ...any) value.Value {
	if configurator := getConfigurator(); configurator != nil {
		return configurator.Get(pattern, def...)
	}
	return value.NewValue()
}

// Set 设置配置值
func Set(pattern string, value any) error {
	if configurator := getConfigurator(); configurator != nil {
		return configurator.Set(pattern, value)
	}
	return nil
}

// Match 匹配多个规则
func Match(patterns ...string) Matcher {
	if configurator := getConfigurator(); configurator != nil {
		return configurator.Match(patterns...)
	}
	return newEmptyMatcher()
}

// Watch 设置监听回调
func Watch(cb WatchCallbackFunc, names ...string) {
	if configurator := getConfigurator(); configurator != nil {
		configurator.Watch(cb, names...)
	}
}

// Load 加载配置项
func Load(ctx context.Context, source string, file ...string) ([]*Configuration, error) {
	if configurator := getConfigurator(); configurator != nil {
		return configurator.Load(ctx, source, file...)
	}
	return nil, nil
}

// Store 保存配置项
func Store(ctx context.Context, source string, file string, content any, override ...bool) error {
	if configurator := getConfigurator(); configurator != nil {
		return configurator.Store(ctx, source, file, content, override...)
	}
	return nil
}

// Close 关闭配置监听
func Close() {
	writeMu.Lock()
	defer writeMu.Unlock()

	if configurator := getGlobalConfigurator(); configurator != nil {
		configurator.Close()
	}
}
