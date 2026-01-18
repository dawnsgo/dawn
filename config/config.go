package config

import (
	"context"
	"sync"

	"github.com/dawnsgo/dawn/core/value"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// Configurator 获取配置器
	Configurator() Configurator
}

var (
	globalConfigurator Configurator
	contextProvider    ContextProvider
	providerMu         sync.RWMutex
)

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

// SetConfigurator 设置配置器
func SetConfigurator(configurator Configurator) {
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalConfigurator != nil {
		globalConfigurator.Close()
	}
	globalConfigurator = configurator
}

// GetConfigurator 获取配置器
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetConfigurator() Configurator {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if configurator := contextProvider.Configurator(); configurator != nil {
			return configurator
		}
	}
	return globalConfigurator
}

// getConfigurator 内部获取配置器（供其他函数调用）
func getConfigurator() Configurator {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if configurator := contextProvider.Configurator(); configurator != nil {
			return configurator
		}
	}
	return globalConfigurator
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
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalConfigurator != nil {
		globalConfigurator.Close()
	}
}
