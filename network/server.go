package network

import "github.com/dawnsgo/dawn/core/buffer"

// 事件处理器类型定义

// StartHandler 服务器启动时的回调函数类型
type StartHandler func()

// CloseHandler 服务器关闭时的回调函数类型
type CloseHandler func()

// ConnectHandler 连接建立时的回调函数类型
type ConnectHandler func(conn Conn)

// DisconnectHandler 连接断开时的回调函数类型
type DisconnectHandler func(conn Conn)

// ReceiveHandler 接收消息时的回调函数类型
type ReceiveHandler func(conn Conn, buf buffer.Buffer)

// Server 网络服务器接口，用于监听和处理客户端连接。
// 实现该接口可以支持多种网络协议（TCP、KCP、WebSocket 等）。
type Server interface {
	// Addr 返回服务器监听的地址
	Addr() string
	// Start 启动服务器，开始监听客户端连接
	Start() error
	// Stop 停止服务器，关闭所有连接
	Stop() error
	// Protocol 返回服务器使用的网络协议名称（如 "tcp"、"kcp"、"ws"）
	Protocol() string
	// OnStart 注册服务器启动时的回调处理器
	OnStart(handler StartHandler)
	// OnStop 注册服务器关闭时的回调处理器
	OnStop(handler CloseHandler)
	// OnConnect 注册连接建立时的回调处理器
	OnConnect(handler ConnectHandler)
	// OnReceive 注册接收到消息时的回调处理器
	OnReceive(handler ReceiveHandler)
	// OnDisconnect 注册连接断开时的回调处理器
	OnDisconnect(handler DisconnectHandler)
}
