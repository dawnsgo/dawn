package dawn

import (
	"sync"

	"github.com/dawnsgo/dawn/cache"
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/eventbus"
	"github.com/dawnsgo/dawn/lock"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/task"
)

// Context 框架上下文，统一管理核心依赖
// 支持多实例场景（如测试），同时保持向后兼容
type Context struct {
	mu           sync.RWMutex
	logger       log.Logger
	configurator config.Configurator
	eventbus     eventbus.Eventbus
	taskPool     task.Pool
	cache        cache.Cache
	lockMaker    lock.Maker
	attached     bool // 是否已关联到各包
}

// defaultContext 默认全局上下文（向后兼容）
var (
	defaultContext     *Context
	defaultContextOnce sync.Once
)

// NewContext 创建新的上下文
func NewContext() *Context {
	return &Context{}
}

// Default 获取默认上下文
func Default() *Context {
	defaultContextOnce.Do(func() {
		defaultContext = NewContext()
		// 默认上下文自动关联到各包
		defaultContext.AttachToPackages()
	})
	return defaultContext
}

// ==================== Context Provider 接口实现 ====================
// 这些方法用于实现各包的 ContextProvider 接口

// Logger 获取日志记录器（实现 log.ContextProvider 接口）
func (c *Context) Logger() log.Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// Configurator 获取配置器（实现 config.ContextProvider 接口）
func (c *Context) Configurator() config.Configurator {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.configurator
}

// Eventbus 获取事件总线（实现 eventbus.ContextProvider 接口）
func (c *Context) Eventbus() eventbus.Eventbus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.eventbus
}

// TaskPool 获取任务池（实现 task.ContextProvider 接口）
func (c *Context) TaskPool() task.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.taskPool
}

// Cache 获取缓存（实现 cache.ContextProvider 接口）
func (c *Context) Cache() cache.Cache {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache
}

// LockMaker 获取分布式锁制造器（实现 lock.ContextProvider 接口）
func (c *Context) LockMaker() lock.Maker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lockMaker
}

// ==================== Setter 方法 ====================

// SetLogger 设置日志记录器
func (c *Context) SetLogger(logger log.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if logger == nil {
		return
	}

	if c.logger != nil {
		c.logger.Close()
	}
	c.logger = logger
}

// SetConfigurator 设置配置器
func (c *Context) SetConfigurator(configurator config.Configurator) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if configurator == nil {
		return
	}

	if c.configurator != nil {
		c.configurator.Close()
	}
	c.configurator = configurator
}

// SetEventbus 设置事件总线
func (c *Context) SetEventbus(eb eventbus.Eventbus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if eb == nil {
		return
	}

	if c.eventbus != nil {
		c.eventbus.Close()
	}
	c.eventbus = eb
}

// SetTaskPool 设置任务池
func (c *Context) SetTaskPool(pool task.Pool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if pool == nil {
		return
	}

	if c.taskPool != nil {
		c.taskPool.Release()
	}
	c.taskPool = pool
}

// SetCache 设置缓存
func (c *Context) SetCache(ca cache.Cache) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ca == nil {
		return
	}

	if c.cache != nil {
		c.cache.Close()
	}
	c.cache = ca
}

// SetLockMaker 设置分布式锁制造器
func (c *Context) SetLockMaker(maker lock.Maker) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if maker == nil {
		return
	}

	if c.lockMaker != nil {
		c.lockMaker.Close()
	}
	c.lockMaker = maker
}

// ==================== Context 关联机制 ====================

// AttachToPackages 将 Context 关联到各包的全局变量
// 调用后，各包的 Get 函数会优先从 Context 读取
func (c *Context) AttachToPackages() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.attached {
		return
	}

	// 将 Context 作为各包的 ContextProvider
	log.SetContextProvider(c)
	config.SetContextProvider(c)
	eventbus.SetContextProvider(c)
	task.SetContextProvider(c)
	cache.SetContextProvider(c)
	lock.SetContextProvider(c)

	c.attached = true
}

// DetachFromPackages 解除 Context 与各包的关联
func (c *Context) DetachFromPackages() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.attached {
		return
	}

	// 解除关联
	log.SetContextProvider(nil)
	config.SetContextProvider(nil)
	eventbus.SetContextProvider(nil)
	task.SetContextProvider(nil)
	cache.SetContextProvider(nil)
	lock.SetContextProvider(nil)

	c.attached = false
}

// IsAttached 检查是否已关联到各包
func (c *Context) IsAttached() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.attached
}

// ==================== Lifecycle ====================

// Close 关闭上下文中的所有资源
func (c *Context) Close() error {
	// 先解除关联
	c.DetachFromPackages()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.eventbus != nil {
		c.eventbus.Close()
		c.eventbus = nil
	}

	if c.lockMaker != nil {
		c.lockMaker.Close()
		c.lockMaker = nil
	}

	if c.cache != nil {
		c.cache.Close()
		c.cache = nil
	}

	if c.taskPool != nil {
		c.taskPool.Release()
		c.taskPool = nil
	}

	if c.configurator != nil {
		c.configurator.Close()
		c.configurator = nil
	}

	if c.logger != nil {
		c.logger.Close()
		c.logger = nil
	}

	return nil
}
