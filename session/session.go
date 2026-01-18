package session

import (
	"net"
	"sync"

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

func (k Kind) String() string {
	switch k {
	case Conn:
		return "conn"
	case User:
		return "user"
	}

	return ""
}

// connShard 连接分片
type connShard struct {
	sync.RWMutex
	items map[int64]network.Conn
}

// userShard 用户分片
type userShard struct {
	sync.RWMutex
	items map[int64]network.Conn
}

// channelShard 频道分片
type channelShard struct {
	sync.RWMutex
	items map[string]map[network.Conn]struct{}
}

// Session 会话管理器（分片锁设计）
// 所有数据结构都使用分片锁，减少锁竞争，提高并发性能
type Session struct {
	shardCount    int             // 分片数量
	connShards    []*connShard    // 连接分片
	userShards    []*userShard    // 用户分片
	channelShards []*channelShard // 频道分片（优化：从全局锁改为分片锁）
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

// NewSession 创建会话管理器
func NewSession(opts ...Option) *Session {
	s := &Session{
		shardCount: defaultShardCount,
	}

	for _, opt := range opts {
		opt(s)
	}

	// 初始化连接分片
	s.connShards = make([]*connShard, s.shardCount)
	for i := 0; i < s.shardCount; i++ {
		s.connShards[i] = &connShard{
			items: make(map[int64]network.Conn),
		}
	}

	// 初始化用户分片
	s.userShards = make([]*userShard, s.shardCount)
	for i := 0; i < s.shardCount; i++ {
		s.userShards[i] = &userShard{
			items: make(map[int64]network.Conn),
		}
	}

	// 初始化频道分片（优化：从全局锁改为分片锁）
	s.channelShards = make([]*channelShard, s.shardCount)
	for i := 0; i < s.shardCount; i++ {
		s.channelShards[i] = &channelShard{
			items: make(map[string]map[network.Conn]struct{}),
		}
	}

	return s
}

// getConnShard 获取连接分片
func (s *Session) getConnShard(cid int64) *connShard {
	return s.connShards[cid%int64(s.shardCount)]
}

// getUserShard 获取用户分片
func (s *Session) getUserShard(uid int64) *userShard {
	return s.userShards[uid%int64(s.shardCount)]
}

// getChannelShard 获取频道分片（使用字符串hash）
func (s *Session) getChannelShard(channel string) *channelShard {
	// 使用简单的字符串hash算法
	var hash uint32
	for i := 0; i < len(channel); i++ {
		hash = hash*31 + uint32(channel[i])
	}
	return s.channelShards[hash%uint32(s.shardCount)]
}

// AddConn 添加连接
func (s *Session) AddConn(conn network.Conn) {
	cid, uid := conn.ID(), conn.UID()

	// 添加到连接分片
	connShard := s.getConnShard(cid)
	connShard.Lock()
	connShard.items[cid] = conn
	connShard.Unlock()

	// 如果已绑定用户，添加到用户分片
	if uid != 0 {
		userShard := s.getUserShard(uid)
		userShard.Lock()
		userShard.items[uid] = conn
		userShard.Unlock()
	}
}

// RemConn 移除连接
func (s *Session) RemConn(conn network.Conn) {
	cid, uid := conn.ID(), conn.UID()

	// 从连接分片移除
	connShard := s.getConnShard(cid)
	connShard.Lock()
	delete(connShard.items, cid)
	connShard.Unlock()

	// 从用户分片移除
	if uid != 0 {
		userShard := s.getUserShard(uid)
		userShard.Lock()
		delete(userShard.items, uid)
		userShard.Unlock()
	}

	// 取消所有频道订阅（使用分片锁）
	conn.Attr().Visit(func(channel, _ any) bool {
		channelName := channel.(string)
		shard := s.getChannelShard(channelName)
		shard.Lock()
		s.doUnsubscribeInShard(shard, channelName, conn)
		shard.Unlock()
		return true
	})
}

// Has 是否存在会话
func (s *Session) Has(kind Kind, target int64) (ok bool, err error) {
	switch kind {
	case Conn:
		shard := s.getConnShard(target)
		shard.RLock()
		_, ok = shard.items[target]
		shard.RUnlock()
	case User:
		shard := s.getUserShard(target)
		shard.RLock()
		_, ok = shard.items[target]
		shard.RUnlock()
	default:
		err = errors.ErrInvalidSessionKind
	}

	return
}

// Bind 绑定用户ID
func (s *Session) Bind(cid, uid int64) error {
	// 获取连接
	connShard := s.getConnShard(cid)
	connShard.RLock()
	conn, ok := connShard.items[cid]
	connShard.RUnlock()

	if !ok {
		return errors.ErrNotFoundSession
	}

	// 处理旧的用户绑定
	if oldUID := conn.UID(); oldUID != 0 {
		if uid == oldUID {
			return nil
		}
		// 移除旧绑定
		oldUserShard := s.getUserShard(oldUID)
		oldUserShard.Lock()
		delete(oldUserShard.items, oldUID)
		oldUserShard.Unlock()
	}

	// 处理新 UID 的旧连接
	userShard := s.getUserShard(uid)
	userShard.Lock()
	if oldConn, exists := userShard.items[uid]; exists {
		oldConn.Unbind()
	}
	conn.Bind(uid)
	userShard.items[uid] = conn
	userShard.Unlock()

	return nil
}

// Unbind 解绑用户ID
func (s *Session) Unbind(uid int64) (int64, error) {
	userShard := s.getUserShard(uid)
	userShard.Lock()
	conn, ok := userShard.items[uid]
	if !ok {
		userShard.Unlock()
		return 0, errors.ErrNotFoundSession
	}

	conn.Unbind()
	delete(userShard.items, uid)
	userShard.Unlock()

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

	switch kind {
	case Conn:
		for _, target := range targets {
			shard := s.getConnShard(target)
			shard.RLock()
			conn, ok := shard.items[target]
			shard.RUnlock()
			if ok && conn.Push(message) == nil {
				n++
			}
		}
	case User:
		for _, target := range targets {
			shard := s.getUserShard(target)
			shard.RLock()
			conn, ok := shard.items[target]
			shard.RUnlock()
			if ok && conn.Push(message) == nil {
				n++
			}
		}
	default:
		err = errors.ErrInvalidSessionKind
	}

	return
}

// Broadcast 推送广播消息（异步）
func (s *Session) Broadcast(kind Kind, message []byte) (n int64, err error) {
	switch kind {
	case Conn:
		for _, shard := range s.connShards {
			shard.RLock()
			for _, conn := range shard.items {
				if conn.Push(message) == nil {
					n++
				}
			}
			shard.RUnlock()
		}
	case User:
		for _, shard := range s.userShards {
			shard.RLock()
			for _, conn := range shard.items {
				if conn.Push(message) == nil {
					n++
				}
			}
			shard.RUnlock()
		}
	default:
		err = errors.ErrInvalidSessionKind
	}

	return
}

// Publish 发布频道消息（异步）
func (s *Session) Publish(channel string, message []byte) (n int64) {
	shard := s.getChannelShard(channel)
	shard.RLock()
	channels, ok := shard.items[channel]
	if !ok {
		shard.RUnlock()
		return
	}

	for conn := range channels {
		if conn.Push(message) == nil {
			n++
		}
	}
	shard.RUnlock()

	return
}

// Subscribe 订阅频道
func (s *Session) Subscribe(kind Kind, targets []int64, channel string) (err error) {
	if len(targets) == 0 {
		return
	}

	// 先收集所有有效连接
	conns := make([]network.Conn, 0, len(targets))

	switch kind {
	case Conn:
		for _, target := range targets {
			shard := s.getConnShard(target)
			shard.RLock()
			if conn, ok := shard.items[target]; ok {
				conns = append(conns, conn)
			}
			shard.RUnlock()
		}
	case User:
		for _, target := range targets {
			shard := s.getUserShard(target)
			shard.RLock()
			if conn, ok := shard.items[target]; ok {
				conns = append(conns, conn)
			}
			shard.RUnlock()
		}
	default:
		err = errors.ErrInvalidSessionKind
		return
	}

	// 订阅频道（使用分片锁）
	shard := s.getChannelShard(channel)
	shard.Lock()
	for _, conn := range conns {
		conn.Attr().Set(channel, struct{}{})

		if channels, ok := shard.items[channel]; ok {
			channels[conn] = struct{}{}
		} else {
			channels = make(map[network.Conn]struct{}, len(targets))
			channels[conn] = struct{}{}
			shard.items[channel] = channels
		}
	}
	shard.Unlock()

	return
}

// Unsubscribe 取消订阅频道
func (s *Session) Unsubscribe(kind Kind, targets []int64, channel string) (err error) {
	if len(targets) == 0 {
		return
	}

	// 先收集所有有效连接
	conns := make([]network.Conn, 0, len(targets))

	switch kind {
	case Conn:
		for _, target := range targets {
			shard := s.getConnShard(target)
			shard.RLock()
			if conn, ok := shard.items[target]; ok {
				conns = append(conns, conn)
			}
			shard.RUnlock()
		}
	case User:
		for _, target := range targets {
			shard := s.getUserShard(target)
			shard.RLock()
			if conn, ok := shard.items[target]; ok {
				conns = append(conns, conn)
			}
			shard.RUnlock()
		}
	default:
		err = errors.ErrInvalidSessionKind
		return
	}

	// 取消订阅（使用分片锁）
	shard := s.getChannelShard(channel)
	shard.Lock()
	for _, conn := range conns {
		if ok := conn.Attr().Del(channel); ok {
			s.doUnsubscribeInShard(shard, channel, conn)
		}
	}
	shard.Unlock()

	return
}

// doUnsubscribeInShard 取消订阅频道（内部方法，调用前需持有分片锁）
func (s *Session) doUnsubscribeInShard(shard *channelShard, channel string, conn network.Conn) {
	if channels, ok := shard.items[channel]; ok {
		delete(channels, conn)

		if len(channels) == 0 {
			delete(shard.items, channel)
		}
	}
}

// Stat 统计会话总数
func (s *Session) Stat(kind Kind) (int64, error) {
	var total int64

	switch kind {
	case Conn:
		for _, shard := range s.connShards {
			shard.RLock()
			total += int64(len(shard.items))
			shard.RUnlock()
		}
	case User:
		for _, shard := range s.userShards {
			shard.RLock()
			total += int64(len(shard.items))
			shard.RUnlock()
		}
	default:
		return 0, errors.ErrInvalidSessionKind
	}

	return total, nil
}

// getConn 获取会话连接
func (s *Session) getConn(kind Kind, target int64) (network.Conn, error) {
	switch kind {
	case Conn:
		shard := s.getConnShard(target)
		shard.RLock()
		conn, ok := shard.items[target]
		shard.RUnlock()
		if !ok {
			return nil, errors.ErrNotFoundSession
		}
		return conn, nil
	case User:
		shard := s.getUserShard(target)
		shard.RLock()
		conn, ok := shard.items[target]
		shard.RUnlock()
		if !ok {
			return nil, errors.ErrNotFoundSession
		}
		return conn, nil
	default:
		return nil, errors.ErrInvalidSessionKind
	}
}
