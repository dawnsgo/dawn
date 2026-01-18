package dawn

import (
	"context"
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

// Container 服务容器
type Container struct {
	ctx        *Context              // 依赖注入上下文
	components []component.Component // 组件列表
}

// ContainerOption 容器配置选项
type ContainerOption func(*Container)

// WithContext 使用指定的上下文
func WithContext(ctx *Context) ContainerOption {
	return func(c *Container) {
		c.ctx = ctx
	}
}

// NewContainer 创建一个容器
func NewContainer(opts ...ContainerOption) *Container {
	c := &Container{
		ctx: Default(), // 默认使用全局上下文
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
	c.doSaveProcessID()

	c.doPrintFrameworkInfo()

	// 确保 Context 已关联到各包
	c.ctx.AttachToPackages()

	if err := c.doInitComponents(); err != nil {
		log.Fatalf("init components failed: %v", err)
	}

	if err := c.doStartComponents(); err != nil {
		log.Fatalf("start components failed: %v", err)
	}

	if len(once) == 0 || !once[0] {
		c.doWaitSystemSignal()
	}

	c.doCloseComponents()

	c.doDestroyComponents()

	c.doClearModules()
}

// 初始化所有组件
func (c *Container) doInitComponents() error {
	for _, comp := range c.components {
		if err := comp.Init(); err != nil {
			return err
		}
	}
	return nil
}

// 启动所有组件
func (c *Container) doStartComponents() error {
	for _, comp := range c.components {
		if err := comp.Start(); err != nil {
			return err
		}
	}
	return nil
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
func (c *Container) doSaveProcessID() {
	filename := etc.Get(defaultPIDKey).String()
	if filename == "" {
		return
	}

	if err := xos.WriteFile(filename, []byte(strconv.Itoa(syscall.Getpid()))); err != nil {
		log.Fatalf("pid save failed: %v", err)
	}
}

// 打印框架信息
func (c *Container) doPrintFrameworkInfo() {
	info.PrintFrameworkInfo()

	info.PrintGlobalInfo()
}
