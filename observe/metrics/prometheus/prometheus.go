/**
 * @Author: dawn
 * @Desc: Prometheus 指标提供者实现
 */

package prometheus

import (
	"github.com/dawnsgo/dawn/observe/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Provider Prometheus 指标提供者
type Provider struct {
	registry *prometheus.Registry
}

// Option 配置选项
type Option func(*Provider)

// WithRegistry 设置注册表
func WithRegistry(registry *prometheus.Registry) Option {
	return func(p *Provider) {
		p.registry = registry
	}
}

// NewProvider 创建 Prometheus 提供者
func NewProvider(opts ...Option) *Provider {
	p := &Provider{
		registry: prometheus.DefaultRegisterer.(*prometheus.Registry),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// NewCounter 创建计数器
func (p *Provider) NewCounter(opts metrics.CounterOpts) metrics.Counter {
	c := promauto.With(p.registry).NewCounter(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	})
	return &counter{c: c}
}

// NewCounterVec 创建带标签的计数器
func (p *Provider) NewCounterVec(opts metrics.CounterOpts, labelNames []string) metrics.CounterVec {
	c := promauto.With(p.registry).NewCounterVec(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, labelNames)
	return &counterVec{c: c}
}

// NewGauge 创建仪表盘
func (p *Provider) NewGauge(opts metrics.GaugeOpts) metrics.Gauge {
	g := promauto.With(p.registry).NewGauge(prometheus.GaugeOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	})
	return &gauge{g: g}
}

// NewGaugeVec 创建带标签的仪表盘
func (p *Provider) NewGaugeVec(opts metrics.GaugeOpts, labelNames []string) metrics.GaugeVec {
	g := promauto.With(p.registry).NewGaugeVec(prometheus.GaugeOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, labelNames)
	return &gaugeVec{g: g}
}

// NewHistogram 创建直方图
func (p *Provider) NewHistogram(opts metrics.HistogramOpts) metrics.Histogram {
	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = metrics.DefaultBuckets
	}
	h := promauto.With(p.registry).NewHistogram(prometheus.HistogramOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   buckets,
	})
	return &histogram{h: h}
}

// NewHistogramVec 创建带标签的直方图
func (p *Provider) NewHistogramVec(opts metrics.HistogramOpts, labelNames []string) metrics.HistogramVec {
	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = metrics.DefaultBuckets
	}
	h := promauto.With(p.registry).NewHistogramVec(prometheus.HistogramOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   buckets,
	}, labelNames)
	return &histogramVec{h: h}
}

// NewSummary 创建摘要
func (p *Provider) NewSummary(opts metrics.SummaryOpts) metrics.Summary {
	objectives := opts.Objectives
	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}
	s := promauto.With(p.registry).NewSummary(prometheus.SummaryOpts{
		Namespace:  opts.Namespace,
		Subsystem:  opts.Subsystem,
		Name:       opts.Name,
		Help:       opts.Help,
		Objectives: objectives,
		MaxAge:     opts.MaxAge,
	})
	return &summary{s: s}
}

// NewSummaryVec 创建带标签的摘要
func (p *Provider) NewSummaryVec(opts metrics.SummaryOpts, labelNames []string) metrics.SummaryVec {
	objectives := opts.Objectives
	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}
	s := promauto.With(p.registry).NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  opts.Namespace,
		Subsystem:  opts.Subsystem,
		Name:       opts.Name,
		Help:       opts.Help,
		Objectives: objectives,
		MaxAge:     opts.MaxAge,
	}, labelNames)
	return &summaryVec{s: s}
}

// ============================================================================
// Counter 适配器
// ============================================================================

type counter struct {
	c prometheus.Counter
}

func (c *counter) Inc()               { c.c.Inc() }
func (c *counter) Add(delta float64)  { c.c.Add(delta) }

type counterVec struct {
	c *prometheus.CounterVec
}

func (v *counterVec) WithLabelValues(lvs ...string) metrics.Counter {
	return &counter{c: v.c.WithLabelValues(lvs...)}
}

// ============================================================================
// Gauge 适配器
// ============================================================================

type gauge struct {
	g prometheus.Gauge
}

func (g *gauge) Set(value float64) { g.g.Set(value) }
func (g *gauge) Inc()              { g.g.Inc() }
func (g *gauge) Dec()              { g.g.Dec() }
func (g *gauge) Add(delta float64) { g.g.Add(delta) }
func (g *gauge) Sub(delta float64) { g.g.Sub(delta) }

type gaugeVec struct {
	g *prometheus.GaugeVec
}

func (v *gaugeVec) WithLabelValues(lvs ...string) metrics.Gauge {
	return &gauge{g: v.g.WithLabelValues(lvs...)}
}

// ============================================================================
// Histogram 适配器
// ============================================================================

type histogram struct {
	h prometheus.Observer
}

func (h *histogram) Observe(value float64) { h.h.Observe(value) }

type histogramVec struct {
	h *prometheus.HistogramVec
}

func (v *histogramVec) WithLabelValues(lvs ...string) metrics.Histogram {
	return &histogram{h: v.h.WithLabelValues(lvs...)}
}

// ============================================================================
// Summary 适配器
// ============================================================================

type summary struct {
	s prometheus.Observer
}

func (s *summary) Observe(value float64) { s.s.Observe(value) }

type summaryVec struct {
	s *prometheus.SummaryVec
}

func (v *summaryVec) WithLabelValues(lvs ...string) metrics.Summary {
	return &summary{s: v.s.WithLabelValues(lvs...)}
}
