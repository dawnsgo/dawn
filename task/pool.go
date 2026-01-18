package task

import (
	"sync"
	"sync/atomic"

	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/utils/xcall"
	"github.com/panjf2000/ants/v2"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// TaskPool 获取任务池
	TaskPool() Pool
}

type Pool interface {
	// AddTask 添加任务
	AddTask(task func()) error
	// Release 释放任务
	Release()
}

var (
	// globalPool 使用 atomic.Value 存储，无锁读取
	globalPool atomic.Value // Pool

	// contextProvider 使用 atomic.Value 存储，无锁读取
	contextProvider atomic.Value // ContextProvider

	// writeMu 保护写入操作的原子性
	writeMu sync.Mutex
)

func init() {
	SetPool(NewPool())
}

type defaultPool struct {
	pool *ants.Pool
}

func NewPool(opts ...Option) *defaultPool {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	p := &defaultPool{}
	p.pool, _ = ants.NewPool(o.size,
		ants.WithLogger(&logger{}),
		ants.WithNonblocking(o.nonblocking),
		ants.WithDisablePurge(o.disablePurge),
	)

	return p
}

// AddTask 添加任务
func (p *defaultPool) AddTask(task func()) error {
	return p.pool.Submit(task)
}

// Release 释放任务
func (p *defaultPool) Release() {
	p.pool.Release()
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

// SetPool 设置任务池
func SetPool(pool Pool) {
	writeMu.Lock()
	defer writeMu.Unlock()

	// 释放旧的 pool
	if old := getGlobalPool(); old != nil {
		old.Release()
	}
	if pool != nil {
		globalPool.Store(pool)
	}
}

// getGlobalPool 获取全局任务池（无锁）
func getGlobalPool() Pool {
	if v := globalPool.Load(); v != nil {
		return v.(Pool)
	}
	return nil
}

// GetPool 获取任务池
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetPool() Pool {
	if provider := GetContextProvider(); provider != nil {
		if pool := provider.TaskPool(); pool != nil {
			return pool
		}
	}
	return getGlobalPool()
}

// getPool 内部获取任务池（供其他函数调用）
func getPool() Pool {
	return GetPool()
}

// AddTask 添加任务
func AddTask(task func()) {
	pool := getPool()
	if pool == nil {
		xcall.Go(task)
		return
	}

	if err := pool.AddTask(task); err != nil {
		xcall.Go(task)
		log.Warnf("add task to the task pool failed: %v", err)
		return
	}
}

// Release 释放任务
func Release() {
	writeMu.Lock()
	defer writeMu.Unlock()

	if pool := getGlobalPool(); pool != nil {
		pool.Release()
	}
}

type logger struct {
}

func (l *logger) Printf(format string, args ...any) {
	log.Infof(format, args...)
}
