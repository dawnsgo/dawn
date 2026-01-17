/**
 * @Author: dawn
 * @Desc: Metrics 全局注册表
 */

package metrics

import (
	"sync"
)

var (
	globalProvider Provider = &noopProvider{}
	providerMu     sync.RWMutex
)

// SetProvider 设置全局指标提供者
func SetProvider(p Provider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	globalProvider = p
}

// GetProvider 获取全局指标提供者
func GetProvider() Provider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return globalProvider
}

// ============================================================================
// 便捷函数
// ============================================================================

// NewCounter 创建计数器
func NewCounter(opts CounterOpts) Counter {
	return GetProvider().NewCounter(opts)
}

// NewCounterVec 创建带标签的计数器
func NewCounterVec(opts CounterOpts, labelNames []string) CounterVec {
	return GetProvider().NewCounterVec(opts, labelNames)
}

// NewGauge 创建仪表盘
func NewGauge(opts GaugeOpts) Gauge {
	return GetProvider().NewGauge(opts)
}

// NewGaugeVec 创建带标签的仪表盘
func NewGaugeVec(opts GaugeOpts, labelNames []string) GaugeVec {
	return GetProvider().NewGaugeVec(opts, labelNames)
}

// NewHistogram 创建直方图
func NewHistogram(opts HistogramOpts) Histogram {
	return GetProvider().NewHistogram(opts)
}

// NewHistogramVec 创建带标签的直方图
func NewHistogramVec(opts HistogramOpts, labelNames []string) HistogramVec {
	return GetProvider().NewHistogramVec(opts, labelNames)
}

// NewSummary 创建摘要
func NewSummary(opts SummaryOpts) Summary {
	return GetProvider().NewSummary(opts)
}

// NewSummaryVec 创建带标签的摘要
func NewSummaryVec(opts SummaryOpts, labelNames []string) SummaryVec {
	return GetProvider().NewSummaryVec(opts, labelNames)
}

// ============================================================================
// Noop 实现（无操作）
// ============================================================================

type noopProvider struct{}

func (p *noopProvider) NewCounter(opts CounterOpts) Counter           { return &noopCounter{} }
func (p *noopProvider) NewCounterVec(opts CounterOpts, labelNames []string) CounterVec {
	return &noopCounterVec{}
}
func (p *noopProvider) NewGauge(opts GaugeOpts) Gauge             { return &noopGauge{} }
func (p *noopProvider) NewGaugeVec(opts GaugeOpts, labelNames []string) GaugeVec {
	return &noopGaugeVec{}
}
func (p *noopProvider) NewHistogram(opts HistogramOpts) Histogram { return &noopHistogram{} }
func (p *noopProvider) NewHistogramVec(opts HistogramOpts, labelNames []string) HistogramVec {
	return &noopHistogramVec{}
}
func (p *noopProvider) NewSummary(opts SummaryOpts) Summary { return &noopSummary{} }
func (p *noopProvider) NewSummaryVec(opts SummaryOpts, labelNames []string) SummaryVec {
	return &noopSummaryVec{}
}

// Noop Counter
type noopCounter struct{}

func (c *noopCounter) Inc()            {}
func (c *noopCounter) Add(delta float64) {}

type noopCounterVec struct{}

func (v *noopCounterVec) WithLabelValues(lvs ...string) Counter { return &noopCounter{} }

// Noop Gauge
type noopGauge struct{}

func (g *noopGauge) Set(value float64) {}
func (g *noopGauge) Inc()              {}
func (g *noopGauge) Dec()              {}
func (g *noopGauge) Add(delta float64) {}
func (g *noopGauge) Sub(delta float64) {}

type noopGaugeVec struct{}

func (v *noopGaugeVec) WithLabelValues(lvs ...string) Gauge { return &noopGauge{} }

// Noop Histogram
type noopHistogram struct{}

func (h *noopHistogram) Observe(value float64) {}

type noopHistogramVec struct{}

func (v *noopHistogramVec) WithLabelValues(lvs ...string) Histogram { return &noopHistogram{} }

// Noop Summary
type noopSummary struct{}

func (s *noopSummary) Observe(value float64) {}

type noopSummaryVec struct{}

func (v *noopSummaryVec) WithLabelValues(lvs ...string) Summary { return &noopSummary{} }
