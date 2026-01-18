// Package network 定义了网络通信的核心接口，包括客户端、服务器和连接。
// 支持 TCP、KCP、WebSocket 等多种网络协议。
package network

// Client 网络客户端接口，用于建立和管理网络连接。
// 实现该接口可以支持多种网络协议（TCP、KCP、WebSocket 等）。
type Client interface {
	// Dial 连接到指定的服务器地址。
	// addr 可以传入多个地址，客户端会依次尝试连接。
	Dial(addr ...string) (Conn, error)
	// Protocol 返回客户端使用的网络协议名称（如 "tcp"、"kcp"、"ws"）。
	Protocol() string
	// OnConnect 注册连接建立时的回调处理器。
	OnConnect(handler ConnectHandler)
	// OnReceive 注册接收到消息时的回调处理器。
	OnReceive(handler ReceiveHandler)
	// OnDisconnect 注册连接断开时的回调处理器。
	OnDisconnect(handler DisconnectHandler)
}
