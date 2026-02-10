/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/7/7 1:19 上午
 * @Desc: 网关服务器
 */

package gate

import (
	"context"
	"fmt"
	"sync"

	"github.com/dawnsgo/dawn/cluster"
	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/component"
	"github.com/dawnsgo/dawn/core/buffer"
	"github.com/dawnsgo/dawn/core/info"
	"github.com/dawnsgo/dawn/core/net"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/internal/transporter/gate"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/network"
	"github.com/dawnsgo/dawn/registry"
	"github.com/dawnsgo/dawn/session"
)

type Gate struct {
	component.Base
	cluster.BaseCluster
	opts     *options
	ctx      context.Context // 用于 NewGate 时 newProxy，Init 时传入 InitBase
	cancel   context.CancelFunc
	proxy    *proxy
	instance *registry.ServiceInstance
	session  *session.Session
	linker   *gate.Server
	wg       *sync.WaitGroup
}

func NewGate(opts ...Option) *Gate {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	g := &Gate{}
	g.opts = o
	g.ctx, g.cancel = context.WithCancel(o.ctx)
	g.proxy = newProxy(g)
	g.session = session.NewSession()
	g.wg = &sync.WaitGroup{}

	return g
}

// Name 组件名称
func (g *Gate) Name() string {
	return g.opts.name
}

// Init 初始化
func (g *Gate) Init() error {
	if g.opts.id == "" {
		return errors.NewWithCode(codes.InvalidArgument, "instance id can not be empty")
	}

	if g.opts.server == nil {
		return errors.NewWithCode(codes.MissingComponent, "server component is not injected")
	}

	if g.opts.locator == nil {
		return errors.NewWithCode(codes.MissingLocator, "locator component is not injected")
	}

	if g.opts.registry == nil {
		return errors.NewWithCode(codes.MissingComponent, "registry component is not injected")
	}

	g.InitBase(g.ctx, g.cancel, g.opts.registry, defaultTimeout)

	return nil
}

// Start 启动组件
func (g *Gate) Start() error {
	if !g.TryStart() {
		return nil
	}

	if err := g.startNetworkServer(); err != nil {
		return err
	}

	if err := g.startLinkerServer(); err != nil {
		return err
	}

	if err := g.registerServiceInstance(); err != nil {
		return err
	}

	if err := g.proxy.watch(); err != nil {
		return err
	}

	g.printInfo()

	return nil
}

// Close 关闭节点
func (g *Gate) Close() error {
	if !g.TryClose() {
		return nil
	}

	g.RefreshInstances()

	g.wg.Wait()

	return nil
}

// Destroy 销毁组件
func (g *Gate) Destroy() error {
	if !g.TryDestroy() {
		return nil
	}

	g.DeregisterInstances()

	g.stopNetworkServer()

	g.stopLinkerServer()

	g.Cancel()

	return nil
}

// 启动网络服务器
func (g *Gate) startNetworkServer() error {
	g.opts.server.OnConnect(g.handleConnect)
	g.opts.server.OnDisconnect(g.handleDisconnect)
	g.opts.server.OnReceive(g.handleReceive)

	if err := g.opts.server.Start(); err != nil {
		return errors.WrapWithCode(err, codes.NetworkError, "network server start failed")
	}

	return nil
}

// 停止网关服务器
func (g *Gate) stopNetworkServer() {
	if err := g.opts.server.Stop(); err != nil {
		log.Errorf("network server stop failed: %v", err)
	}
}

// 处理连接打开
func (g *Gate) handleConnect(conn network.Conn) {
	g.wg.Add(1)

	g.session.AddConn(conn)

	cid, uid := conn.ID(), conn.UID()

	ctx, cancel := context.WithTimeout(g.Context(), g.opts.timeout)
	g.proxy.trigger(ctx, cluster.Connect, cid, uid)
	cancel()
}

// 处理断开连接
func (g *Gate) handleDisconnect(conn network.Conn) {
	g.session.RemConn(conn)

	if cid, uid := conn.ID(), conn.UID(); uid != 0 {
		ctx, cancel := context.WithTimeout(g.Context(), g.opts.timeout)
		_ = g.proxy.unbindGate(ctx, cid, uid)
		g.proxy.trigger(ctx, cluster.Disconnect, cid, uid)
		cancel()
	} else {
		ctx, cancel := context.WithTimeout(g.Context(), g.opts.timeout)
		g.proxy.trigger(ctx, cluster.Disconnect, cid, uid)
		cancel()
	}

	g.wg.Done()
}

// 处理接收到的消息
func (g *Gate) handleReceive(conn network.Conn, buf buffer.Buffer) {
	cid, uid := conn.ID(), conn.UID()
	ctx, cancel := context.WithTimeout(g.Context(), g.opts.timeout)
	g.proxy.deliver(ctx, cid, uid, buf)
	cancel()
}

// 启动传输服务器
func (g *Gate) startLinkerServer() error {
	transporter, err := gate.NewServer(&provider{gate: g}, &gate.ServerOptions{
		Addr:   g.opts.addr,
		Expose: g.opts.expose,
	})
	if err != nil {
		return errors.WrapWithCode(err, codes.InternalError, "link server create failed")
	}

	g.linker = transporter

	go func() {
		if err = g.linker.Start(); err != nil {
			log.Errorf("link server start failed: %v", err)
		}
	}()

	return nil
}

// 停止传输服务器
func (g *Gate) stopLinkerServer() {
	if err := g.linker.Stop(); err != nil {
		log.Errorf("link server stop failed: %v", err)
	}
}

// 注册服务实例
func (g *Gate) registerServiceInstance() error {
	g.instance = &registry.ServiceInstance{
		ID:       g.opts.id,
		Name:     cluster.Gate.String(),
		Kind:     cluster.Gate.String(),
		Alias:    g.opts.name,
		State:    g.GetState().String(),
		Endpoint: g.linker.Endpoint().String(),
		Metadata: g.opts.metadata,
	}

	g.ClearInstances()
	g.AddInstance(g.instance)

	return g.RegisterInstances()
}

// 打印组件信息
func (g *Gate) printInfo() {
	infos := make([]string, 0, 6)
	infos = append(infos, fmt.Sprintf("ID: %s", g.opts.id))
	infos = append(infos, fmt.Sprintf("Name: %s", g.Name()))
	infos = append(infos, fmt.Sprintf("Link: %s", g.linker.ExposeAddr()))
	infos = append(infos, fmt.Sprintf("Server: [%s] %s", g.opts.server.Protocol(), net.FulfillAddr(g.opts.server.Addr())))
	infos = append(infos, fmt.Sprintf("Locator: %s", g.opts.locator.Name()))
	infos = append(infos, fmt.Sprintf("Registry: %s", g.opts.registry.Name()))

	info.PrintBoxInfo("Gate", infos...)
}
