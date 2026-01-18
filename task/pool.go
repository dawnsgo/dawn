package task

import (
	"sync"

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
	globalPool      Pool
	contextProvider ContextProvider
	providerMu      sync.RWMutex
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

// SetPool 设置任务池
func SetPool(pool Pool) {
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalPool != nil {
		globalPool.Release()
	}
	globalPool = pool
}

// GetPool 获取任务池
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetPool() Pool {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if pool := contextProvider.TaskPool(); pool != nil {
			return pool
		}
	}
	return globalPool
}

// getPool 内部获取任务池（供其他函数调用）
func getPool() Pool {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if pool := contextProvider.TaskPool(); pool != nil {
			return pool
		}
	}
	return globalPool
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
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalPool != nil {
		globalPool.Release()
	}
}

type logger struct {
}

func (l *logger) Printf(format string, args ...any) {
	log.Infof(format, args...)
}
