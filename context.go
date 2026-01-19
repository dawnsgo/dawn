package dawn

import (
	"sync"
	"sync/atomic"

	"github.com/dawnsgo/dawn/cache"
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/eventbus"
	"github.com/dawnsgo/dawn/lock"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/task"
)

// Context 框架上下文，统一管理核心依赖
// 支持多实例场景（如测试），同时保持向后兼容
//
// 并发安全设计：
//   - Getter 方法使用 atomic.Value，无锁读取，高性能
//   - Setter 方法使用 sync.Mutex，保证写入原子性
//   - 适合读多写少的场景（Setter 只在初始化时调用）
type Context struct {
	// 使用 atomic.Value 存储各依赖项，避免读锁开销
	// atomic.Value 适合读多写少的场景
	// 使用包装器结构体存储，以支持存储 nil 值
	logger       atomic.Value // *loggerWrapper
	configurator atomic.Value // *configuratorWrapper
	eventbus     atomic.Value // *eventbusWrapper
	taskPool     atomic.Value // *taskPoolWrapper
	cache        atomic.Value // *cacheWrapper
	lockMaker    atomic.Value // *lockMakerWrapper

	// 写锁，保护 Setter 方法的原子性（关闭旧资源 + 设置新资源）
	writeMu sync.Mutex

	// attached 状态使用 atomic.Bool
	attached atomic.Bool
}

// ==================== 包装器类型 ====================
// 使用包装器解决 atomic.Value 不能存储 nil 的问题

type loggerWrapper struct{ v log.Logger }
type configuratorWrapper struct{ v config.Configurator }
type eventbusWrapper struct{ v eventbus.Eventbus }
type taskPoolWrapper struct{ v task.Pool }
type cacheWrapper struct{ v cache.Cache }
type lockMakerWrapper struct{ v lock.Maker }

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
// 使用 atomic.Load 无锁读取，高性能

// Logger 获取日志记录器（实现 log.ContextProvider 接口）
func (c *Context) Logger() log.Logger {
	if v := c.logger.Load(); v != nil {
		return v.(*loggerWrapper).v
	}
	return nil
}

// Configurator 获取配置器（实现 config.ContextProvider 接口）
func (c *Context) Configurator() config.Configurator {
	if v := c.configurator.Load(); v != nil {
		return v.(*configuratorWrapper).v
	}
	return nil
}

// Eventbus 获取事件总线（实现 eventbus.ContextProvider 接口）
func (c *Context) Eventbus() eventbus.Eventbus {
	if v := c.eventbus.Load(); v != nil {
		return v.(*eventbusWrapper).v
	}
	return nil
}

// TaskPool 获取任务池（实现 task.ContextProvider 接口）
func (c *Context) TaskPool() task.Pool {
	if v := c.taskPool.Load(); v != nil {
		return v.(*taskPoolWrapper).v
	}
	return nil
}

// Cache 获取缓存（实现 cache.ContextProvider 接口）
func (c *Context) Cache() cache.Cache {
	if v := c.cache.Load(); v != nil {
		return v.(*cacheWrapper).v
	}
	return nil
}

// LockMaker 获取分布式锁制造器（实现 lock.ContextProvider 接口）
func (c *Context) LockMaker() lock.Maker {
	if v := c.lockMaker.Load(); v != nil {
		return v.(*lockMakerWrapper).v
	}
	return nil
}

// ==================== Setter 方法 ====================
// Setter 方法使用 Mutex 保护，确保"关闭旧资源+设置新资源"的原子性

// SetLogger 设置日志记录器
func (c *Context) SetLogger(logger log.Logger) {
	if logger == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 关闭旧的 logger
	if old := c.Logger(); old != nil {
		old.Close()
	}
	c.logger.Store(&loggerWrapper{v: logger})
}

// SetConfigurator 设置配置器
func (c *Context) SetConfigurator(configurator config.Configurator) {
	if configurator == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 关闭旧的 configurator
	if old := c.Configurator(); old != nil {
		old.Close()
	}
	c.configurator.Store(&configuratorWrapper{v: configurator})
}

// SetEventbus 设置事件总线
func (c *Context) SetEventbus(eb eventbus.Eventbus) {
	if eb == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 关闭旧的 eventbus
	if old := c.Eventbus(); old != nil {
		old.Close()
	}
	c.eventbus.Store(&eventbusWrapper{v: eb})
}

// SetTaskPool 设置任务池
func (c *Context) SetTaskPool(pool task.Pool) {
	if pool == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 释放旧的 taskPool
	if old := c.TaskPool(); old != nil {
		old.Release()
	}
	c.taskPool.Store(&taskPoolWrapper{v: pool})
}

// SetCache 设置缓存
func (c *Context) SetCache(ca cache.Cache) {
	if ca == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 关闭旧的 cache
	if old := c.Cache(); old != nil {
		old.Close()
	}
	c.cache.Store(&cacheWrapper{v: ca})
}

// SetLockMaker 设置分布式锁制造器
func (c *Context) SetLockMaker(maker lock.Maker) {
	if maker == nil {
		return
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 关闭旧的 lockMaker
	if old := c.LockMaker(); old != nil {
		old.Close()
	}
	c.lockMaker.Store(&lockMakerWrapper{v: maker})
}

// ==================== Context 关联机制 ====================

// AttachToPackages 将 Context 关联到各包的全局变量
// 调用后，各包的 Get 函数会优先从 Context 读取
func (c *Context) AttachToPackages() {
	// 使用 CAS 避免重复关联
	if !c.attached.CompareAndSwap(false, true) {
		return
	}

	// 将 Context 作为各包的 ContextProvider
	log.SetContextProvider(c)
	config.SetContextProvider(c)
	eventbus.SetContextProvider(c)
	task.SetContextProvider(c)
	cache.SetContextProvider(c)
	lock.SetContextProvider(c)
}

// DetachFromPackages 解除 Context 与各包的关联
func (c *Context) DetachFromPackages() {
	// 使用 CAS 避免重复解除
	if !c.attached.CompareAndSwap(true, false) {
		return
	}

	// 解除关联
	log.SetContextProvider(nil)
	config.SetContextProvider(nil)
	eventbus.SetContextProvider(nil)
	task.SetContextProvider(nil)
	cache.SetContextProvider(nil)
	lock.SetContextProvider(nil)
}

// IsAttached 检查是否已关联到各包
func (c *Context) IsAttached() bool {
	return c.attached.Load()
}

// ==================== Lifecycle ====================

// Close 关闭上下文中的所有资源
// 关闭顺序按照依赖关系设计：
// 1. Eventbus - 先关闭事件总线，停止消息发布
// 2. LockMaker - 关闭分布式锁，释放锁资源
// 3. Cache - 关闭缓存连接
// 4. TaskPool - 等待任务完成并释放任务池
// 5. Configurator - 关闭配置监听
// 6. Logger - 最后关闭日志，确保其他组件的日志能够输出
func (c *Context) Close() error {
	// 先解除关联
	c.DetachFromPackages()

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	// 按依赖关系顺序关闭资源
	// 使用包装器存储 nil 值，避免 atomic.Value 存储 nil 的问题

	// 1. 关闭事件总线（可能依赖日志）
	if eb := c.Eventbus(); eb != nil {
		eb.Close()
		c.eventbus.Store(&eventbusWrapper{v: nil})
	}

	// 2. 关闭分布式锁（可能依赖日志）
	if lm := c.LockMaker(); lm != nil {
		lm.Close()
		c.lockMaker.Store(&lockMakerWrapper{v: nil})
	}

	// 3. 关闭缓存（可能依赖日志）
	if ca := c.Cache(); ca != nil {
		ca.Close()
		c.cache.Store(&cacheWrapper{v: nil})
	}

	// 4. 释放任务池（等待任务完成）
	if tp := c.TaskPool(); tp != nil {
		tp.Release()
		c.taskPool.Store(&taskPoolWrapper{v: nil})
	}

	// 5. 关闭配置器（可能触发配置变更回调）
	if cfg := c.Configurator(); cfg != nil {
		cfg.Close()
		c.configurator.Store(&configuratorWrapper{v: nil})
	}

	// 6. 最后关闭日志（确保所有组件的日志都能输出）
	if lg := c.Logger(); lg != nil {
		lg.Close()
		c.logger.Store(&loggerWrapper{v: nil})
	}

	return nil
}
