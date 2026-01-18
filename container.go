package dawn

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/dawnsgo/dawn/component"
	"github.com/dawnsgo/dawn/core/info"
	"github.com/dawnsgo/dawn/etc"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/utils/xcall"
	"github.com/dawnsgo/dawn/utils/xos"
)

const (
	defaultPIDKey                 = "etc.pid"                 // 进程文件路径
	defaultShutdownMaxWaitTimeKey = "etc.shutdownMaxWaitTime" // 容器关闭最大等待时间
)

// ContainerError 容器错误类型
type ContainerError struct {
	Phase     string // 错误发生的阶段: init, start, close, destroy
	Component string // 出错的组件名称
	Err       error  // 原始错误
}

func (e *ContainerError) Error() string {
	return fmt.Sprintf("[%s] component [%s] failed: %v", e.Phase, e.Component, e.Err)
}

func (e *ContainerError) Unwrap() error {
	return e.Err
}

// ErrorHandler 错误处理器类型
type ErrorHandler func(err error)

// Container 服务容器
type Container struct {
	ctx              *Context              // 依赖注入上下文
	components       []component.Component // 组件列表
	initializedComps []component.Component // 已初始化的组件（用于回滚）
	startedComps     []component.Component // 已启动的组件（用于回滚）
	errorHandler     ErrorHandler          // 自定义错误处理器
	exitOnError      bool                  // 发生错误时是否退出程序
}

// ContainerOption 容器配置选项
type ContainerOption func(*Container)

// WithContext 使用指定的上下文
func WithContext(ctx *Context) ContainerOption {
	return func(c *Container) {
		c.ctx = ctx
	}
}

// WithErrorHandler 设置自定义错误处理器
func WithErrorHandler(handler ErrorHandler) ContainerOption {
	return func(c *Container) {
		c.errorHandler = handler
	}
}

// WithExitOnError 设置发生错误时是否退出程序（默认 true）
func WithExitOnError(exit bool) ContainerOption {
	return func(c *Container) {
		c.exitOnError = exit
	}
}

// NewContainer 创建一个容器
func NewContainer(opts ...ContainerOption) *Container {
	c := &Container{
		ctx:         Default(), // 默认使用全局上下文
		exitOnError: true,      // 默认发生错误时退出
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Context 获取容器的上下文
func (c *Container) Context() *Context {
	return c.ctx
}

// Add 添加组件
func (c *Container) Add(components ...component.Component) {
	c.components = append(c.components, components...)
}

// Serve 启动容器
func (c *Container) Serve(once ...bool) {
	if err := c.doSaveProcessID(); err != nil {
		c.handleError(&ContainerError{Phase: "setup", Component: "pid", Err: err})
		return
	}

	c.doPrintFrameworkInfo()

	// 确保 Context 已关联到各包
	c.ctx.AttachToPackages()

	if err := c.doInitComponents(); err != nil {
		c.handleError(err)
		// 回滚已初始化的组件
		c.rollbackInit()
		return
	}

	if err := c.doStartComponents(); err != nil {
		c.handleError(err)
		// 回滚：先关闭已启动的组件，再销毁已初始化的组件
		c.rollbackStart()
		c.rollbackInit()
		return
	}

	if len(once) == 0 || !once[0] {
		c.doWaitSystemSignal()
	}

	c.doCloseComponents()

	c.doDestroyComponents()

	c.doClearModules()
}

// ServeWithError 启动容器并返回错误（不会调用 log.Fatal）
// 适用于测试或需要自定义错误处理的场景
func (c *Container) ServeWithError(once ...bool) error {
	if err := c.doSaveProcessID(); err != nil {
		return &ContainerError{Phase: "setup", Component: "pid", Err: err}
	}

	c.doPrintFrameworkInfo()

	// 确保 Context 已关联到各包
	c.ctx.AttachToPackages()

	if err := c.doInitComponents(); err != nil {
		// 回滚已初始化的组件
		c.rollbackInit()
		return err
	}

	if err := c.doStartComponents(); err != nil {
		// 回滚：先关闭已启动的组件，再销毁已初始化的组件
		c.rollbackStart()
		c.rollbackInit()
		return err
	}

	if len(once) == 0 || !once[0] {
		c.doWaitSystemSignal()
	}

	c.doCloseComponents()

	c.doDestroyComponents()

	c.doClearModules()

	return nil
}

// handleError 处理错误
func (c *Container) handleError(err error) {
	// 如果有自定义错误处理器，先调用它
	if c.errorHandler != nil {
		c.errorHandler(err)
	}

	// 记录错误日志
	log.Errorf("container error: %v", err)

	// 根据配置决定是否退出
	if c.exitOnError {
		log.Fatalf("container fatal error: %v", err)
	}
}

// 初始化所有组件（带回滚支持）
func (c *Container) doInitComponents() error {
	c.initializedComps = make([]component.Component, 0, len(c.components))

	for _, comp := range c.components {
		if err := comp.Init(); err != nil {
			return &ContainerError{
				Phase:     "init",
				Component: comp.Name(),
				Err:       err,
			}
		}
		// 记录已初始化的组件
		c.initializedComps = append(c.initializedComps, comp)
	}
	return nil
}

// 启动所有组件（带回滚支持）
func (c *Container) doStartComponents() error {
	c.startedComps = make([]component.Component, 0, len(c.components))

	for _, comp := range c.components {
		if err := comp.Start(); err != nil {
			return &ContainerError{
				Phase:     "start",
				Component: comp.Name(),
				Err:       err,
			}
		}
		// 记录已启动的组件
		c.startedComps = append(c.startedComps, comp)
	}
	return nil
}

// rollbackInit 回滚已初始化的组件（逆序销毁）
func (c *Container) rollbackInit() {
	if len(c.initializedComps) == 0 {
		return
	}

	log.Warnf("rolling back %d initialized components...", len(c.initializedComps))

	// 逆序销毁已初始化的组件
	for i := len(c.initializedComps) - 1; i >= 0; i-- {
		comp := c.initializedComps[i]
		if err := comp.Destroy(); err != nil {
			log.Warnf("rollback destroy component [%s] failed: %v", comp.Name(), err)
		} else {
			log.Infof("rollback destroyed component [%s]", comp.Name())
		}
	}

	c.initializedComps = nil
}

// rollbackStart 回滚已启动的组件（逆序关闭）
func (c *Container) rollbackStart() {
	if len(c.startedComps) == 0 {
		return
	}

	log.Warnf("rolling back %d started components...", len(c.startedComps))

	// 逆序关闭已启动的组件
	for i := len(c.startedComps) - 1; i >= 0; i-- {
		comp := c.startedComps[i]
		if err := comp.Close(); err != nil {
			log.Warnf("rollback close component [%s] failed: %v", comp.Name(), err)
		} else {
			log.Infof("rollback closed component [%s]", comp.Name())
		}
	}

	c.startedComps = nil
}

// 关闭所有组件
func (c *Container) doCloseComponents() {
	g := xcall.NewGoroutines()

	for _, comp := range c.components {
		comp := comp // 避免闭包捕获问题
		g.Add(func() {
			if err := comp.Close(); err != nil {
				log.Warnf("close component [%s] failed: %v", comp.Name(), err)
			}
		})
	}

	g.Run(context.Background(), etc.Get(defaultShutdownMaxWaitTimeKey).Duration())
}

// 销毁所有组件
func (c *Container) doDestroyComponents() {
	g := xcall.NewGoroutines()

	for _, comp := range c.components {
		comp := comp // 避免闭包捕获问题
		g.Add(func() {
			if err := comp.Destroy(); err != nil {
				log.Warnf("destroy component [%s] failed: %v", comp.Name(), err)
			}
		})
	}

	g.Run(context.Background(), 5*time.Second)
}

// 等待系统信号
func (c *Container) doWaitSystemSignal() {
	sig := make(chan os.Signal, 1)

	switch runtime.GOOS {
	case `windows`:
		signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	default:
		signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGTERM)
	}

	s := <-sig

	signal.Stop(sig)

	log.Warnf("process got signal %v, container will close", s)
}

// 清理所有模块
// 由于现在使用 Context 作为唯一数据源，只需要关闭 Context 即可
func (c *Container) doClearModules() {
	// 关闭 etc（这个不受 Context 管理）
	etc.Close()

	// 关闭 Context，会自动解除关联并关闭所有资源
	if c.ctx != nil {
		c.ctx.Close()
	}
}

// 保存进程号
func (c *Container) doSaveProcessID() error {
	filename := etc.Get(defaultPIDKey).String()
	if filename == "" {
		return nil
	}

	if err := xos.WriteFile(filename, []byte(strconv.Itoa(syscall.Getpid()))); err != nil {
		return err
	}
	return nil
}

// 打印框架信息
func (c *Container) doPrintFrameworkInfo() {
	info.PrintFrameworkInfo()

	info.PrintGlobalInfo()
}
