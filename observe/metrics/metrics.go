/**
 * @Author: dawn
 * @Desc: Metrics 抽象层定义
 */

package metrics

import (
	"time"
)

// ============================================================================
// 指标类型定义
// ============================================================================

// Counter 计数器接口（只增不减）
type Counter interface {
	// Inc 增加 1
	Inc()
	// Add 增加指定值（必须 >= 0）
	Add(delta float64)
}

// Gauge 仪表盘接口（可增可减）
type Gauge interface {
	// Set 设置值
	Set(value float64)
	// Inc 增加 1
	Inc()
	// Dec 减少 1
	Dec()
	// Add 增加指定值
	Add(delta float64)
	// Sub 减少指定值
	Sub(delta float64)
}

// Histogram 直方图接口（用于观测值分布）
type Histogram interface {
	// Observe 观测一个值
	Observe(value float64)
}

// Summary 摘要接口（用于计算分位数）
type Summary interface {
	// Observe 观测一个值
	Observe(value float64)
}

// Timer 计时器辅助接口
type Timer interface {
	// ObserveDuration 观测持续时间
	ObserveDuration()
}

// ============================================================================
// 带标签的指标类型
// ============================================================================

// CounterVec 带标签的计数器
type CounterVec interface {
	// WithLabelValues 根据标签值获取计数器
	WithLabelValues(lvs ...string) Counter
}

// GaugeVec 带标签的仪表盘
type GaugeVec interface {
	// WithLabelValues 根据标签值获取仪表盘
	WithLabelValues(lvs ...string) Gauge
}

// HistogramVec 带标签的直方图
type HistogramVec interface {
	// WithLabelValues 根据标签值获取直方图
	WithLabelValues(lvs ...string) Histogram
}

// SummaryVec 带标签的摘要
type SummaryVec interface {
	// WithLabelValues 根据标签值获取摘要
	WithLabelValues(lvs ...string) Summary
}

// ============================================================================
// Provider 接口
// ============================================================================

// Provider 指标提供者接口
type Provider interface {
	// NewCounter 创建计数器
	NewCounter(opts CounterOpts) Counter
	// NewCounterVec 创建带标签的计数器
	NewCounterVec(opts CounterOpts, labelNames []string) CounterVec

	// NewGauge 创建仪表盘
	NewGauge(opts GaugeOpts) Gauge
	// NewGaugeVec 创建带标签的仪表盘
	NewGaugeVec(opts GaugeOpts, labelNames []string) GaugeVec

	// NewHistogram 创建直方图
	NewHistogram(opts HistogramOpts) Histogram
	// NewHistogramVec 创建带标签的直方图
	NewHistogramVec(opts HistogramOpts, labelNames []string) HistogramVec

	// NewSummary 创建摘要
	NewSummary(opts SummaryOpts) Summary
	// NewSummaryVec 创建带标签的摘要
	NewSummaryVec(opts SummaryOpts, labelNames []string) SummaryVec
}

// ============================================================================
// 指标选项
// ============================================================================

// CounterOpts 计数器选项
type CounterOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

// GaugeOpts 仪表盘选项
type GaugeOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
}

// HistogramOpts 直方图选项
type HistogramOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Buckets   []float64 // 桶边界
}

// SummaryOpts 摘要选项
type SummaryOpts struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	Objectives map[float64]float64 // 分位数目标
	MaxAge     time.Duration       // 观测值最大存活时间
}

// ============================================================================
// 默认桶边界
// ============================================================================

// DefaultBuckets 默认延迟桶边界（毫秒）
var DefaultBuckets = []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// DefaultByteBuckets 默认字节大小桶边界
var DefaultByteBuckets = []float64{64, 256, 512, 1024, 4096, 16384, 65536, 262144, 1048576}

// ============================================================================
// 辅助函数
// ============================================================================

// NewTimer 创建计时器
func NewTimer(h Histogram) Timer {
	return &timer{
		histogram: h,
		start:     time.Now(),
	}
}

type timer struct {
	histogram Histogram
	start     time.Time
}

func (t *timer) ObserveDuration() {
	if t.histogram != nil {
		t.histogram.Observe(float64(time.Since(t.start).Milliseconds()))
	}
}
