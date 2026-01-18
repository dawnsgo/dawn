/**
 * @Author: dawn
 * @Desc: 统一的心跳检测机制，为各网络协议提供复用的心跳检测功能
 */

package network

import (
	"sync/atomic"
	"time"

	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/utils/xtime"
)

// HeartbeatMechanism 心跳机制类型
type HeartbeatMechanism string

const (
	// RespHeartbeat 响应式心跳：收到客户端心跳后响应
	RespHeartbeat HeartbeatMechanism = "resp"
	// TickHeartbeat 主动定时心跳：服务端主动发送心跳
	TickHeartbeat HeartbeatMechanism = "tick"
)

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	Interval  time.Duration      // 心跳检测间隔时间
	Mechanism HeartbeatMechanism // 心跳机制
	Timeout   time.Duration      // 心跳超时时间（默认为 2 * Interval）
}

// DefaultHeartbeatConfig 默认心跳配置
func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		Interval:  10 * time.Second,
		Mechanism: RespHeartbeat,
		Timeout:   0, // 使用默认值 2 * Interval
	}
}

// GetTimeout 获取超时时间
func (c *HeartbeatConfig) GetTimeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 2 * c.Interval
}

// HeartbeatHandler 心跳处理回调
type HeartbeatHandler struct {
	// OnTimeout 心跳超时回调
	OnTimeout func()
	// OnSendHeartbeat 发送心跳回调
	OnSendHeartbeat func() error
}

// HeartbeatChecker 心跳检测器
type HeartbeatChecker struct {
	config            *HeartbeatConfig
	handler           *HeartbeatHandler
	lastHeartbeatTime atomic.Int64
	connID            int64
	stopped           atomic.Bool
}

// NewHeartbeatChecker 创建心跳检测器
func NewHeartbeatChecker(config *HeartbeatConfig, handler *HeartbeatHandler, connID int64) *HeartbeatChecker {
	hc := &HeartbeatChecker{
		config:  config,
		handler: handler,
		connID:  connID,
	}
	hc.UpdateHeartbeat()
	return hc
}

// UpdateHeartbeat 更新最后心跳时间
func (hc *HeartbeatChecker) UpdateHeartbeat() {
	hc.lastHeartbeatTime.Store(xtime.Now().UnixNano())
}

// GetLastHeartbeatTime 获取最后心跳时间
func (hc *HeartbeatChecker) GetLastHeartbeatTime() int64 {
	return hc.lastHeartbeatTime.Load()
}

// IsEnabled 检查心跳是否启用
func (hc *HeartbeatChecker) IsEnabled() bool {
	return hc.config != nil && hc.config.Interval > 0
}

// Stop 停止心跳检测
func (hc *HeartbeatChecker) Stop() {
	hc.stopped.Store(true)
}

// IsStopped 检查是否已停止
func (hc *HeartbeatChecker) IsStopped() bool {
	return hc.stopped.Load()
}

// CheckTimeout 检查心跳是否超时
// 返回 true 表示超时
func (hc *HeartbeatChecker) CheckTimeout(currentTime time.Time) bool {
	if !hc.IsEnabled() || hc.IsStopped() {
		return false
	}

	deadline := currentTime.Add(-hc.config.GetTimeout()).UnixNano()
	return hc.lastHeartbeatTime.Load() < deadline
}

// HandleTick 处理心跳定时器tick
// 返回 false 表示连接应该关闭
func (hc *HeartbeatChecker) HandleTick(currentTime time.Time) bool {
	if !hc.IsEnabled() || hc.IsStopped() {
		return true
	}

	// 检查超时
	if hc.CheckTimeout(currentTime) {
		log.Debugf("connection heartbeat timeout, cid: %d", hc.connID)
		if hc.handler != nil && hc.handler.OnTimeout != nil {
			hc.handler.OnTimeout()
		}
		return false
	}

	// 主动心跳模式下发送心跳
	if hc.config.Mechanism == TickHeartbeat {
		if hc.handler != nil && hc.handler.OnSendHeartbeat != nil {
			if err := hc.handler.OnSendHeartbeat(); err != nil {
				log.Errorf("send heartbeat failed, cid: %d, err: %v", hc.connID, err)
			}
		}
	}

	return true
}

// HandleReceivedHeartbeat 处理收到的心跳包
// shouldRespond 返回是否应该响应心跳
func (hc *HeartbeatChecker) HandleReceivedHeartbeat() (shouldRespond bool) {
	hc.UpdateHeartbeat()
	return hc.config.Mechanism == RespHeartbeat
}

// StartTicker 启动心跳定时器
// 返回 ticker 和 stop channel
func (hc *HeartbeatChecker) StartTicker() (*time.Ticker, chan struct{}) {
	if !hc.IsEnabled() {
		return nil, nil
	}

	ticker := time.NewTicker(hc.config.Interval)
	stopCh := make(chan struct{})

	return ticker, stopCh
}

// HeartbeatStats 心跳统计信息
type HeartbeatStats struct {
	ConnID            int64
	LastHeartbeatTime time.Time
	Interval          time.Duration
	Mechanism         HeartbeatMechanism
	IsTimeout         bool
}

// GetStats 获取心跳统计信息
func (hc *HeartbeatChecker) GetStats() *HeartbeatStats {
	lastTime := hc.lastHeartbeatTime.Load()
	return &HeartbeatStats{
		ConnID:            hc.connID,
		LastHeartbeatTime: time.Unix(0, lastTime),
		Interval:          hc.config.Interval,
		Mechanism:         hc.config.Mechanism,
		IsTimeout:         hc.CheckTimeout(xtime.Now()),
	}
}
