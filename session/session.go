package session

import (
	"net"

	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/network"
)

const (
	Conn Kind = iota + 1 // 连接SESSION
	User                 // 用户SESSION
)

// 默认分片数量
const defaultShardCount = 32

type Kind int

// IsValid 检查 Kind 是否有效
func (k Kind) IsValid() bool {
	return k == Conn || k == User
}

func (k Kind) String() string {
	switch k {
	case Conn:
		return "conn"
	case User:
		return "user"
	}

	return ""
}

// Session 会话管理器（泛型 ShardMap 分片锁设计）
// 使用 ShardMap[K,V] 统一分片存储，减少锁竞争
type Session struct {
	shardCount int
	connMap    *ShardMap[int64, network.Conn]           // 连接分片
	userMap    *ShardMap[int64, network.Conn]           // 用户分片
	channelMap *ShardMap[string, map[network.Conn]struct{}] // 频道分片
}

// Option 配置选项
type Option func(*Session)

// WithShardCount 设置分片数量
func WithShardCount(count int) Option {
	return func(s *Session) {
		if count > 0 {
			s.shardCount = count
		}
	}
}

// NewSession 创建会话管理器（内部使用泛型 ShardMap）
func NewSession(opts ...Option) *Session {
	s := &Session{
		shardCount: defaultShardCount,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.connMap = NewInt64ShardMap[network.Conn](s.shardCount)
	s.userMap = NewInt64ShardMap[network.Conn](s.shardCount)
	s.channelMap = NewStringShardMap[map[network.Conn]struct{}](s.shardCount)

	return s
}

// validateKind 验证 Kind 参数是否有效
func (s *Session) validateKind(kind Kind) error {
	if !kind.IsValid() {
		return errors.ErrInvalidSessionKind
	}
	return nil
}

// collectConns 根据 Kind 和目标列表收集连接
func (s *Session) collectConns(kind Kind, targets []int64) ([]network.Conn, error) {
	if err := s.validateKind(kind); err != nil {
		return nil, err
	}

	conns := make([]network.Conn, 0, len(targets))
	switch kind {
	case Conn:
		for _, target := range targets {
			if conn, ok := s.connMap.Get(target); ok {
				conns = append(conns, conn)
			}
		}
	case User:
		for _, target := range targets {
			if conn, ok := s.userMap.Get(target); ok {
				conns = append(conns, conn)
			}
		}
	}
	return conns, nil
}

// AddConn 添加连接
func (s *Session) AddConn(conn network.Conn) {
	cid, uid := conn.ID(), conn.UID()

	s.connMap.Set(cid, conn)
	if uid != 0 {
		s.userMap.Set(uid, conn)
	}
}

// RemConn 移除连接
func (s *Session) RemConn(conn network.Conn) {
	cid, uid := conn.ID(), conn.UID()

	s.connMap.Delete(cid)
	if uid != 0 {
		s.userMap.Delete(uid)
	}

	conn.Attr().Visit(func(channel, _ any) bool {
		channelName := channel.(string)
		s.channelMap.WithShard(channelName, func(items map[string]map[network.Conn]struct{}) {
			s.doUnsubscribeInShardMap(items, channelName, conn)
		})
		return true
	})
}

// Has 是否存在会话
func (s *Session) Has(kind Kind, target int64) (ok bool, err error) {
	if err = s.validateKind(kind); err != nil {
		return
	}

	switch kind {
	case Conn:
		ok = s.connMap.Has(target)
	case User:
		ok = s.userMap.Has(target)
	}

	return
}

// Bind 绑定用户ID
func (s *Session) Bind(cid, uid int64) error {
	conn, ok := s.connMap.Get(cid)
	if !ok {
		return errors.ErrNotFoundSession
	}

	if oldUID := conn.UID(); oldUID != 0 {
		if uid == oldUID {
			return nil
		}
		s.userMap.Delete(oldUID)
	}

	s.userMap.WithShard(uid, func(items map[int64]network.Conn) {
		if oldConn, exists := items[uid]; exists {
			oldConn.Unbind()
		}
		conn.Bind(uid)
		items[uid] = conn
	})

	return nil
}

// Unbind 解绑用户ID
func (s *Session) Unbind(uid int64) (int64, error) {
	conn, ok := s.userMap.GetAndDelete(uid)
	if !ok {
		return 0, errors.ErrNotFoundSession
	}
	conn.Unbind()
	return conn.ID(), nil
}

// LocalIP 获取本地IP
func (s *Session) LocalIP(kind Kind, target int64) (string, error) {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return "", err
	}

	return conn.LocalIP()
}

// LocalAddr 获取本地地址
func (s *Session) LocalAddr(kind Kind, target int64) (net.Addr, error) {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return nil, err
	}

	return conn.LocalAddr()
}

// RemoteIP 获取远端IP
func (s *Session) RemoteIP(kind Kind, target int64) (string, error) {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return "", err
	}

	return conn.RemoteIP()
}

// RemoteAddr 获取远端地址
func (s *Session) RemoteAddr(kind Kind, target int64) (net.Addr, error) {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return nil, err
	}

	return conn.RemoteAddr()
}

// Close 关闭会话
func (s *Session) Close(kind Kind, target int64, force ...bool) error {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return err
	}

	return conn.Close(force...)
}

// Send 发送消息（同步）
func (s *Session) Send(kind Kind, target int64, message []byte) error {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return err
	}

	return conn.Send(message)
}

// Push 推送消息（异步）
func (s *Session) Push(kind Kind, target int64, message []byte) error {
	conn, err := s.getConn(kind, target)
	if err != nil {
		return err
	}

	return conn.Push(message)
}

// Multicast 推送组播消息（异步）
func (s *Session) Multicast(kind Kind, targets []int64, message []byte) (n int64, err error) {
	if len(targets) == 0 {
		return
	}

	if err = s.validateKind(kind); err != nil {
		return
	}

	switch kind {
	case Conn:
		for _, target := range targets {
			if conn, ok := s.connMap.Get(target); ok && conn.Push(message) == nil {
				n++
			}
		}
	case User:
		for _, target := range targets {
			if conn, ok := s.userMap.Get(target); ok && conn.Push(message) == nil {
				n++
			}
		}
	}

	return
}

// Broadcast 推送广播消息（异步）
func (s *Session) Broadcast(kind Kind, message []byte) (n int64, err error) {
	if err = s.validateKind(kind); err != nil {
		return
	}

	switch kind {
	case Conn:
		s.connMap.RangeAll(func(_ int64, conn network.Conn) bool {
			if conn.Push(message) == nil {
				n++
			}
			return true
		})
	case User:
		s.userMap.RangeAll(func(_ int64, conn network.Conn) bool {
			if conn.Push(message) == nil {
				n++
			}
			return true
		})
	}

	return
}

// Publish 发布频道消息（异步）
func (s *Session) Publish(channel string, message []byte) (n int64) {
	s.channelMap.WithShardRLock(channel, func(items map[string]map[network.Conn]struct{}) {
		channels, ok := items[channel]
		if !ok {
			return
		}
		for conn := range channels {
			if conn.Push(message) == nil {
				n++
			}
		}
	})
	return
}

// Subscribe 订阅频道
func (s *Session) Subscribe(kind Kind, targets []int64, channel string) (err error) {
	if len(targets) == 0 {
		return
	}

	// 使用辅助方法收集连接
	conns, err := s.collectConns(kind, targets)
	if err != nil {
		return err
	}

	s.channelMap.WithShard(channel, func(items map[string]map[network.Conn]struct{}) {
		for _, conn := range conns {
			conn.Attr().Set(channel, struct{}{})
			if channels, ok := items[channel]; ok {
				channels[conn] = struct{}{}
			} else {
				channels = make(map[network.Conn]struct{}, len(targets))
				channels[conn] = struct{}{}
				items[channel] = channels
			}
		}
	})

	return
}

// Unsubscribe 取消订阅频道
func (s *Session) Unsubscribe(kind Kind, targets []int64, channel string) (err error) {
	if len(targets) == 0 {
		return
	}

	// 使用辅助方法收集连接
	conns, err := s.collectConns(kind, targets)
	if err != nil {
		return err
	}

	s.channelMap.WithShard(channel, func(items map[string]map[network.Conn]struct{}) {
		for _, conn := range conns {
			if conn.Attr().Del(channel) {
				s.doUnsubscribeInShardMap(items, channel, conn)
			}
		}
	})

	return
}

// doUnsubscribeInShardMap 取消订阅频道（在 WithShard 回调内调用，items 为当前分片 map）
func (s *Session) doUnsubscribeInShardMap(items map[string]map[network.Conn]struct{}, channel string, conn network.Conn) {
	if channels, ok := items[channel]; ok {
		delete(channels, conn)
		if len(channels) == 0 {
			delete(items, channel)
		}
	}
}

// Stat 统计会话总数
func (s *Session) Stat(kind Kind) (int64, error) {
	if err := s.validateKind(kind); err != nil {
		return 0, err
	}

	switch kind {
	case Conn:
		return s.connMap.Len(), nil
	case User:
		return s.userMap.Len(), nil
	}
	return 0, nil
}

// getConn 获取会话连接
func (s *Session) getConn(kind Kind, target int64) (network.Conn, error) {
	if err := s.validateKind(kind); err != nil {
		return nil, err
	}

	switch kind {
	case Conn:
		if conn, ok := s.connMap.Get(target); ok {
			return conn, nil
		}
		return nil, errors.ErrNotFoundSession
	case User:
		if conn, ok := s.userMap.Get(target); ok {
			return conn, nil
		}
		return nil, errors.ErrNotFoundSession
	}

	return nil, errors.ErrInvalidSessionKind
}
