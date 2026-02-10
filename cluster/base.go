package cluster

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/registry"
	"github.com/dawnsgo/dawn/utils/xcall"
	"golang.org/x/sync/errgroup"
)

const (
	// DefaultRegistryTimeout 默认注册/注销超时时间
	DefaultRegistryTimeout = 3 * time.Second
)

// BaseCluster 集群组件公共基类
// 提取 Gate/Node/Mesh 共享的状态管理、服务注册/注销、钩子函数等逻辑
type BaseCluster struct {
	state     atomic.Int32
	ctx       context.Context
	cancel    context.CancelFunc
	instances []*registry.ServiceInstance
	registry  registry.Registry
	timeout   time.Duration
	rw        sync.RWMutex
	hooks     map[Hook][]func()
}

// InitBase 初始化基类
func (b *BaseCluster) InitBase(ctx context.Context, cancel context.CancelFunc, reg registry.Registry, timeout time.Duration) {
	b.ctx = ctx
	b.cancel = cancel
	b.registry = reg
	b.timeout = timeout
	b.hooks = make(map[Hook][]func())
	b.instances = make([]*registry.ServiceInstance, 0)
	b.state.Store(int32(Shut))
}

// ==================== 状态管理 ====================

// GetState 获取状态
func (b *BaseCluster) GetState() State {
	return State(b.state.Load())
}

// SetState 设置状态
func (b *BaseCluster) SetState(state State) {
	b.state.Store(int32(state))
}

// CompareAndSwapState CAS 设置状态
func (b *BaseCluster) CompareAndSwapState(old, new State) bool {
	return b.state.CompareAndSwap(int32(old), int32(new))
}

// TryStart 尝试从 Shut 切换到 Work（启动）
func (b *BaseCluster) TryStart() bool {
	return b.CompareAndSwapState(Shut, Work)
}

// TryClose 尝试切换到 Hang（关闭）
func (b *BaseCluster) TryClose() bool {
	if b.CompareAndSwapState(Work, Hang) {
		return true
	}
	return b.CompareAndSwapState(Busy, Hang)
}

// TryDestroy 尝试从 Hang 切换到 Shut（销毁）
func (b *BaseCluster) TryDestroy() bool {
	return b.CompareAndSwapState(Hang, Shut)
}

// Context 获取上下文
func (b *BaseCluster) Context() context.Context {
	return b.ctx
}

// Cancel 取消上下文
func (b *BaseCluster) Cancel() {
	b.cancel()
}

// ==================== 服务注册 ====================

// AddInstance 添加服务实例
func (b *BaseCluster) AddInstance(instance *registry.ServiceInstance) {
	b.instances = append(b.instances, instance)
}

// ClearInstances 清空服务实例列表（用于 Start 前重置，如 Gate/Mesh 单实例场景）
func (b *BaseCluster) ClearInstances() {
	if b.instances != nil {
		b.instances = b.instances[:0]
	}
}

// RegisterInstances 注册所有服务实例
func (b *BaseCluster) RegisterInstances() error {
	if len(b.instances) == 0 {
		return nil
	}

	eg, ctx := errgroup.WithContext(b.ctx)

	for i := range b.instances {
		instance := b.instances[i]
		eg.Go(func() error {
			tctx, tcancel := context.WithTimeout(ctx, b.timeout)
			defer tcancel()
			return b.registry.Register(tctx, instance)
		})
	}

	if err := eg.Wait(); err != nil {
		return errors.WrapWithCode(err, codes.ServiceRegisterFailed, "register cluster instances failed")
	}

	return nil
}

// RefreshInstances 刷新所有服务实例状态
func (b *BaseCluster) RefreshInstances() {
	state := b.GetState().String()
	for _, instance := range b.instances {
		instance.State = state
	}

	if err := b.doRegisterInstances(); err != nil {
		log.Errorf("refresh cluster instances failed: %v", err)
	}
}

// DeregisterInstances 解注册所有服务实例
func (b *BaseCluster) DeregisterInstances() {
	eg, ctx := errgroup.WithContext(b.ctx)
	for i := range b.instances {
		instance := b.instances[i]
		eg.Go(func() error {
			tctx, tcancel := context.WithTimeout(ctx, b.timeout)
			defer tcancel()
			return b.registry.Deregister(tctx, instance)
		})
	}

	if err := eg.Wait(); err != nil {
		log.Errorf("deregister cluster instances failed: %v", err)
	}
}

func (b *BaseCluster) doRegisterInstances() error {
	eg, ctx := errgroup.WithContext(b.ctx)
	for i := range b.instances {
		instance := b.instances[i]
		eg.Go(func() error {
			tctx, tcancel := context.WithTimeout(ctx, b.timeout)
			defer tcancel()
			return b.registry.Register(tctx, instance)
		})
	}
	return eg.Wait()
}

// ==================== 钩子管理 ====================

// AddHook 添加钩子监听器（线程安全）
func (b *BaseCluster) AddHook(hook Hook, handler func()) {
	switch hook {
	case Destroy:
		// Destroy 钩子允许在任何时候添加
		b.rw.Lock()
		b.hooks[hook] = append(b.hooks[hook], handler)
		b.rw.Unlock()
	default:
		if b.GetState() == Shut {
			b.hooks[hook] = append(b.hooks[hook], handler)
		} else {
			log.Warnf("server is working, can't add hook handler")
		}
	}
}

// RunHooks 执行钩子函数
func (b *BaseCluster) RunHooks(hook Hook) {
	b.rw.RLock()

	if handlers, ok := b.hooks[hook]; ok {
		wg := &sync.WaitGroup{}
		wg.Add(len(handlers))

		for i := range handlers {
			handler := handlers[i]
			xcall.Go(func() {
				handler()
				wg.Done()
			})
		}

		b.rw.RUnlock()
		wg.Wait()
	} else {
		b.rw.RUnlock()
	}
}
