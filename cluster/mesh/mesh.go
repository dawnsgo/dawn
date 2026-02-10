package mesh

import (
	"context"
	"fmt"

	"github.com/dawnsgo/dawn/cluster"
	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/component"
	"github.com/dawnsgo/dawn/core/info"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/registry"
	"github.com/dawnsgo/dawn/transport"
)

type HookHandler func(proxy *Proxy)

type Mesh struct {
	component.Base
	cluster.BaseCluster
	opts        *options
	proxy       *Proxy
	transporter transport.Server
	services    []*serviceEntity
	instance    *registry.ServiceInstance
}

type serviceEntity struct {
	name     string // 服务名称;用于定位服务发现
	desc     any    // 服务描述(grpc为desc描述对象; rpcx为服务路径)
	provider any    // 服务提供者
}

func NewMesh(opts ...Option) *Mesh {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	m := &Mesh{}
	m.opts = o
	m.services = make([]*serviceEntity, 0)
	m.proxy = newProxy(m)

	return m
}

// Name 组件名称
func (m *Mesh) Name() string {
	return m.opts.name
}

// Init 初始化节点
func (m *Mesh) Init() error {
	if m.opts.codec == nil {
		return errors.NewWithCode(codes.MissingComponent, "codec component is not injected")
	}

	if m.opts.registry == nil {
		return errors.NewWithCode(codes.MissingComponent, "registry component is not injected")
	}

	if m.opts.transporter == nil {
		return errors.NewWithCode(codes.MissingTransporter, "transporter component is not injected")
	}

	ctx, cancel := context.WithCancel(m.opts.ctx)
	m.InitBase(ctx, cancel, m.opts.registry, defaultTimeout)

	m.runHookFunc(cluster.Init)

	return nil
}

// Start 启动
func (m *Mesh) Start() error {
	if !m.TryStart() {
		return nil
	}

	if err := m.startTransportServer(); err != nil {
		return err
	}

	if err := m.registerServiceInstance(); err != nil {
		return err
	}

	if err := m.proxy.watch(); err != nil {
		return err
	}

	m.printInfo()

	m.runHookFunc(cluster.Start)

	return nil
}

// Close 关闭
func (m *Mesh) Close() error {
	if !m.TryClose() {
		return nil
	}

	m.RefreshInstances()

	m.runHookFunc(cluster.Close)

	return nil
}

// Destroy 销毁
func (m *Mesh) Destroy() error {
	if !m.TryDestroy() {
		return nil
	}

	m.runHookFunc(cluster.Destroy)

	m.DeregisterInstances()

	m.stopTransportServer()

	m.Cancel()

	return nil
}

// Proxy 获取节点代理
func (m *Mesh) Proxy() *Proxy {
	return m.proxy
}

// 启动传输服务器
func (m *Mesh) startTransportServer() error {
	m.opts.transporter.SetDefaultDiscovery(m.opts.registry)

	transporter, err := m.opts.transporter.NewServer()
	if err != nil {
		return errors.WrapWithCode(err, codes.InternalError, "transport server create failed")
	}

	m.transporter = transporter

	for _, entity := range m.services {
		if err = m.transporter.RegisterService(entity.desc, entity.provider); err != nil {
			return errors.WrapWithCode(err, codes.ServiceRegisterFailed, "register service failed")
		}
	}

	go func() {
		if err = m.transporter.Start(); err != nil {
			log.Errorf("transport server start failed: %v", err)
		}
	}()

	return nil
}

// 停止传输服务器
func (m *Mesh) stopTransportServer() {
	if err := m.transporter.Stop(); err != nil {
		log.Errorf("transport server stop failed: %v", err)
	}
}

// 注册服务实例
func (m *Mesh) registerServiceInstance() error {
	m.instance = &registry.ServiceInstance{
		ID:       m.opts.id,
		Name:     cluster.Mesh.String(),
		Kind:     cluster.Mesh.String(),
		Alias:    m.opts.name,
		State:    m.GetState().String(),
		Endpoint: m.transporter.Endpoint().String(),
		Services: make([]string, 0, len(m.services)),
		Metadata: m.opts.metadata,
	}

	for _, item := range m.services {
		m.instance.Services = append(m.instance.Services, item.name)
	}

	m.ClearInstances()
	m.AddInstance(m.instance)

	return m.RegisterInstances()
}

// 获取状态
func (m *Mesh) getState() cluster.State {
	return m.GetState()
}

// 执行钩子函数（委托 BaseCluster.RunHooks）
func (m *Mesh) runHookFunc(hook cluster.Hook) {
	m.RunHooks(hook)
}

// 添加钩子监听器（包装为 func() 委托 BaseCluster）
func (m *Mesh) addHookListener(hook cluster.Hook, handler HookHandler) {
	m.AddHook(hook, func() { handler(m.proxy) })
}

// 添加服务提供者
func (m *Mesh) addServiceProvider(name string, desc, provider any) {
	if m.getState() == cluster.Shut {
		m.services = append(m.services, &serviceEntity{
			name:     name,
			desc:     desc,
			provider: provider,
		})
	} else {
		log.Warnf("mesh server is working, can't add service provider")
	}
}

// 打印组件信息
func (m *Mesh) printInfo() {
	infos := make([]string, 0, 7)
	infos = append(infos, fmt.Sprintf("ID: %s", m.opts.id))
	infos = append(infos, fmt.Sprintf("Name: %s", m.Name()))
	infos = append(infos, fmt.Sprintf("Codec: %s", m.opts.codec.Name()))

	if m.opts.locator != nil {
		infos = append(infos, fmt.Sprintf("Locator: %s", m.opts.locator.Name()))
	} else {
		infos = append(infos, "Locator: -")
	}

	infos = append(infos, fmt.Sprintf("Registry: %s", m.opts.registry.Name()))

	if m.opts.encryptor != nil {
		infos = append(infos, fmt.Sprintf("Encryptor: %s", m.opts.encryptor.Name()))
	} else {
		infos = append(infos, "Encryptor: -")
	}

	infos = append(infos, fmt.Sprintf("Transporter: %s", m.opts.transporter.Name()))

	info.PrintBoxInfo("Mesh", infos...)
}
