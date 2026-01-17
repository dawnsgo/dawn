/**
 * @Author: dawn
 * @Desc: Session 单元测试
 */

package session

import (
	"net"
	"testing"

	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/network"
)

// mockConn 模拟连接
type mockConn struct {
	id      int64
	uid     int64
	localIP string
	attrs   network.Attr
	state   network.ConnState
}

func (m *mockConn) ID() int64                    { return m.id }
func (m *mockConn) UID() int64                   { return m.uid }
func (m *mockConn) Bind(uid int64)               { m.uid = uid }
func (m *mockConn) Unbind()                      { m.uid = 0 }
func (m *mockConn) State() network.ConnState     { return m.state }
func (m *mockConn) LocalIP() (string, error)     { return m.localIP, nil }
func (m *mockConn) LocalAddr() (net.Addr, error) { return nil, nil }
func (m *mockConn) RemoteIP() (string, error)    { return "", nil }
func (m *mockConn) RemoteAddr() (net.Addr, error) { return nil, nil }
func (m *mockConn) Send([]byte) error            { return nil }
func (m *mockConn) Push([]byte) error             { return nil }
func (m *mockConn) Close(...bool) error          { return nil }
func (m *mockConn) Attr() network.Attr {
	if m.attrs == nil {
		m.attrs = network.NewAttr()
	}
	return m.attrs
}

func TestNewSession(t *testing.T) {
	// 测试默认创建
	s := NewSession()
	if s == nil {
		t.Fatal("NewSession() returned nil")
	}
	if s.shardCount != defaultShardCount {
		t.Fatalf("Expected shardCount %d, got %d", defaultShardCount, s.shardCount)
	}
	if len(s.connShards) != defaultShardCount {
		t.Fatalf("Expected %d conn shards, got %d", defaultShardCount, len(s.connShards))
	}
	if len(s.userShards) != defaultShardCount {
		t.Fatalf("Expected %d user shards, got %d", defaultShardCount, len(s.userShards))
	}

	// 测试自定义分片数量
	customCount := 16
	s2 := NewSession(WithShardCount(customCount))
	if s2.shardCount != customCount {
		t.Fatalf("Expected shardCount %d, got %d", customCount, s2.shardCount)
	}

	// 测试无效分片数量（应该使用默认值）
	s3 := NewSession(WithShardCount(0))
	if s3.shardCount != defaultShardCount {
		t.Fatalf("Expected shardCount %d for invalid count, got %d", defaultShardCount, s3.shardCount)
	}
}

func TestKind_String(t *testing.T) {
	if Conn.String() != "conn" {
		t.Fatalf("Expected 'conn', got '%s'", Conn.String())
	}
	if User.String() != "user" {
		t.Fatalf("Expected 'user', got '%s'", User.String())
	}
	if Kind(99).String() != "" {
		t.Fatal("Invalid Kind should return empty string")
	}
}

func TestSession_AddConn(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}

	s.AddConn(conn)

	// 验证连接已添加
	ok, err := s.Has(Conn, 1)
	if err != nil {
		t.Fatalf("Has() failed: %v", err)
	}
	if !ok {
		t.Fatal("Connection not found after AddConn()")
	}

	// 测试添加已绑定用户的连接
	conn2 := &mockConn{id: 2, uid: 100}
	s.AddConn(conn2)

	ok, err = s.Has(User, 100)
	if err != nil {
		t.Fatalf("Has() failed: %v", err)
	}
	if !ok {
		t.Fatal("User session not found after AddConn() with UID")
	}
}

func TestSession_RemConn(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 100}
	s.AddConn(conn)

	// 验证连接存在
	ok, _ := s.Has(Conn, 1)
	if !ok {
		t.Fatal("Connection should exist")
	}

	// 移除连接
	s.RemConn(conn)

	// 验证连接已移除
	ok, _ = s.Has(Conn, 1)
	if ok {
		t.Fatal("Connection should be removed")
	}

	// 验证用户会话也已移除
	ok, _ = s.Has(User, 100)
	if ok {
		t.Fatal("User session should be removed")
	}
}

func TestSession_Has(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 100}
	s.AddConn(conn)

	// 测试连接存在
	ok, err := s.Has(Conn, 1)
	if err != nil {
		t.Fatalf("Has() failed: %v", err)
	}
	if !ok {
		t.Fatal("Connection should exist")
	}

	// 测试用户存在
	ok, err = s.Has(User, 100)
	if err != nil {
		t.Fatalf("Has() failed: %v", err)
	}
	if !ok {
		t.Fatal("User should exist")
	}

	// 测试不存在的连接
	ok, err = s.Has(Conn, 999)
	if err != nil {
		t.Fatalf("Has() failed: %v", err)
	}
	if ok {
		t.Fatal("Connection should not exist")
	}

	// 测试无效类型
	_, err = s.Has(Kind(99), 1)
	if err == nil {
		t.Fatal("Expected error for invalid Kind")
	}
	if !errors.IsErr(err, errors.ErrInvalidSessionKind) {
		t.Fatalf("Expected ErrInvalidSessionKind, got %v", err)
	}
}

func TestSession_Bind(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	// 测试绑定
	err := s.Bind(1, 100)
	if err != nil {
		t.Fatalf("Bind() failed: %v", err)
	}

	// 验证用户会话已创建
	ok, err := s.Has(User, 100)
	if err != nil {
		t.Fatalf("Has() failed: %v", err)
	}
	if !ok {
		t.Fatal("User session should exist after Bind()")
	}

	// 测试绑定不存在的连接
	err = s.Bind(999, 200)
	if err == nil {
		t.Fatal("Expected error for non-existent connection")
	}
	if !errors.IsErr(err, errors.ErrNotFoundSession) {
		t.Fatalf("Expected ErrNotFoundSession, got %v", err)
	}

	// 测试重复绑定相同 UID
	err = s.Bind(1, 100)
	if err != nil {
		t.Fatalf("Re-binding same UID should succeed: %v", err)
	}
}

func TestSession_Unbind(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 100}
	s.AddConn(conn)

	// 测试解绑
	cid, err := s.Unbind(100)
	if err != nil {
		t.Fatalf("Unbind() failed: %v", err)
	}
	if cid != 1 {
		t.Fatalf("Expected connection ID 1, got %d", cid)
	}

	// 验证用户会话已移除
	ok, _ := s.Has(User, 100)
	if ok {
		t.Fatal("User session should be removed after Unbind()")
	}

	// 测试解绑不存在的用户
	_, err = s.Unbind(999)
	if err == nil {
		t.Fatal("Expected error for non-existent user")
	}
	if !errors.IsErr(err, errors.ErrNotFoundSession) {
		t.Fatalf("Expected ErrNotFoundSession, got %v", err)
	}
}

func TestSession_GetConn(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 100}
	s.AddConn(conn)

	// 测试通过连接 ID 获取
	gotConn, err := s.getConn(Conn, 1)
	if err != nil {
		t.Fatalf("getConn() failed: %v", err)
	}
	if gotConn.ID() != conn.ID() {
		t.Fatal("getConn() returned wrong connection")
	}

	// 测试通过用户 ID 获取
	gotConn, err = s.getConn(User, 100)
	if err != nil {
		t.Fatalf("getConn() failed: %v", err)
	}
	if gotConn.ID() != conn.ID() {
		t.Fatal("getConn() returned wrong connection")
	}

	// 测试不存在的连接
	_, err = s.getConn(Conn, 999)
	if err == nil {
		t.Fatal("Expected error for non-existent connection")
	}

	// 测试无效类型
	_, err = s.getConn(Kind(99), 1)
	if err == nil {
		t.Fatal("Expected error for invalid Kind")
	}
}

func TestSession_LocalIP(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0, localIP: "127.0.0.1"}
	s.AddConn(conn)

	ip, err := s.LocalIP(Conn, 1)
	if err != nil {
		t.Fatalf("LocalIP() failed: %v", err)
	}
	if ip != "127.0.0.1" {
		t.Fatalf("Expected '127.0.0.1', got '%s'", ip)
	}
}

func TestSession_Send(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	err := s.Send(Conn, 1, []byte("test"))
	if err != nil {
		t.Fatalf("Send() failed: %v", err)
	}
}

func TestSession_Push(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	err := s.Push(Conn, 1, []byte("test"))
	if err != nil {
		t.Fatalf("Push() failed: %v", err)
	}
}

func TestSession_Multicast(t *testing.T) {
	s := NewSession()
	conn1 := &mockConn{id: 1, uid: 0}
	conn2 := &mockConn{id: 2, uid: 0}
	s.AddConn(conn1)
	s.AddConn(conn2)

	// 测试组播
	n, err := s.Multicast(Conn, []int64{1, 2}, []byte("test"))
	if err != nil {
		t.Fatalf("Multicast() failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("Expected 2, got %d", n)
	}

	// 测试空列表
	n, err = s.Multicast(Conn, []int64{}, []byte("test"))
	if err != nil {
		t.Fatalf("Multicast() failed: %v", err)
	}
	if n != 0 {
		t.Fatalf("Expected 0, got %d", n)
	}

	// 测试用户组播
	conn3 := &mockConn{id: 3, uid: 100}
	conn4 := &mockConn{id: 4, uid: 200}
	s.AddConn(conn3)
	s.AddConn(conn4)
	s.Bind(3, 100)
	s.Bind(4, 200)

	n, err = s.Multicast(User, []int64{100, 200}, []byte("test"))
	if err != nil {
		t.Fatalf("Multicast() failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("Expected 2, got %d", n)
	}

	// 测试无效类型
	_, err = s.Multicast(Kind(99), []int64{1}, []byte("test"))
	if err == nil {
		t.Fatal("Expected error for invalid Kind")
	}
}

func TestSession_Broadcast(t *testing.T) {
	s := NewSession()
	conn1 := &mockConn{id: 1, uid: 0}
	conn2 := &mockConn{id: 2, uid: 0}
	s.AddConn(conn1)
	s.AddConn(conn2)

	// 测试连接广播
	n, err := s.Broadcast(Conn, []byte("test"))
	if err != nil {
		t.Fatalf("Broadcast() failed: %v", err)
	}
	if n != 2 {
		t.Fatalf("Expected 2, got %d", n)
	}

	// 测试用户广播
	conn3 := &mockConn{id: 3, uid: 100}
	s.AddConn(conn3)
	s.Bind(3, 100)

	n, err = s.Broadcast(User, []byte("test"))
	if err != nil {
		t.Fatalf("Broadcast() failed: %v", err)
	}
	if n != 1 {
		t.Fatalf("Expected 1, got %d", n)
	}

	// 测试无效类型
	_, err = s.Broadcast(Kind(99), []byte("test"))
	if err == nil {
		t.Fatal("Expected error for invalid Kind")
	}
}

func TestSession_Subscribe(t *testing.T) {
	s := NewSession()
	conn1 := &mockConn{id: 1, uid: 0}
	s.AddConn(conn1)

	// 测试订阅
	err := s.Subscribe(Conn, []int64{1}, "channel1")
	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// 测试发布
	n := s.Publish("channel1", []byte("test"))
	if n != 1 {
		t.Fatalf("Expected 1, got %d", n)
	}

	// 测试空列表
	err = s.Subscribe(Conn, []int64{}, "channel2")
	if err != nil {
		t.Fatalf("Subscribe() with empty list should succeed: %v", err)
	}

	// 测试无效类型
	err = s.Subscribe(Kind(99), []int64{1}, "channel3")
	if err == nil {
		t.Fatal("Expected error for invalid Kind")
	}
}

func TestSession_Unsubscribe(t *testing.T) {
	s := NewSession()
	conn1 := &mockConn{id: 1, uid: 0}
	s.AddConn(conn1)

	// 先订阅
	err := s.Subscribe(Conn, []int64{1}, "channel1")
	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// 测试取消订阅
	err = s.Unsubscribe(Conn, []int64{1}, "channel1")
	if err != nil {
		t.Fatalf("Unsubscribe() failed: %v", err)
	}

	// 验证发布不会发送到已取消订阅的连接
	n := s.Publish("channel1", []byte("test"))
	if n != 0 {
		t.Fatalf("Expected 0, got %d", n)
	}

	// 测试空列表
	err = s.Unsubscribe(Conn, []int64{}, "channel2")
	if err != nil {
		t.Fatalf("Unsubscribe() with empty list should succeed: %v", err)
	}

	// 测试无效类型
	err = s.Unsubscribe(Kind(99), []int64{1}, "channel3")
	if err == nil {
		t.Fatal("Expected error for invalid Kind")
	}
}

func TestSession_RemoteIP(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	_, err := s.RemoteIP(Conn, 1)
	if err != nil {
		t.Fatalf("RemoteIP() failed: %v", err)
	}
}

func TestSession_RemoteAddr(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	_, err := s.RemoteAddr(Conn, 1)
	if err != nil {
		t.Fatalf("RemoteAddr() failed: %v", err)
	}
}

func TestSession_LocalAddr(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	_, err := s.LocalAddr(Conn, 1)
	if err != nil {
		t.Fatalf("LocalAddr() failed: %v", err)
	}
}

func TestSession_Close(t *testing.T) {
	s := NewSession()
	conn := &mockConn{id: 1, uid: 0}
	s.AddConn(conn)

	err := s.Close(Conn, 1)
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
}
