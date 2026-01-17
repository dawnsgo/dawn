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
	mu          sync.RWMutex
	logger      log.Logger
	configurator config.Configurator
	eventbus    eventbus.Eventbus
	taskPool    task.Pool
	cache       cache.Cache
	lockMaker   lock.Maker
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
	})
	return defaultContext
}

// ==================== Logger ====================

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

// Logger 获取日志记录器
func (c *Context) Logger() log.Logger {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logger
}

// ==================== Configurator ====================

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

// Configurator 获取配置器
func (c *Context) Configurator() config.Configurator {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.configurator
}

// ==================== Eventbus ====================

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

// Eventbus 获取事件总线
func (c *Context) Eventbus() eventbus.Eventbus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.eventbus
}

// ==================== TaskPool ====================

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

// TaskPool 获取任务池
func (c *Context) TaskPool() task.Pool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.taskPool
}

// ==================== Cache ====================

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

// Cache 获取缓存
func (c *Context) Cache() cache.Cache {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache
}

// ==================== LockMaker ====================

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

// LockMaker 获取分布式锁制造器
func (c *Context) LockMaker() lock.Maker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lockMaker
}

// ==================== Lifecycle ====================

// Close 关闭上下文中的所有资源
func (c *Context) Close() error {
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
