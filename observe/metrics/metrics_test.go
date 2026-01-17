/**
 * @Author: dawn
 * @Desc: Metrics 单元测试
 */

package metrics

import (
	"testing"
)

func TestNewCounter(t *testing.T) {
	// 使用 noop provider
	opts := CounterOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "counter",
		Help:      "test counter",
	}

	counter := NewCounter(opts)
	if counter == nil {
		t.Fatal("NewCounter() returned nil")
	}

	// 测试方法不会 panic
	counter.Inc()
	counter.Add(1.5)
	counter.Add(0)
}

func TestNewGauge(t *testing.T) {
	opts := GaugeOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "gauge",
		Help:      "test gauge",
	}

	gauge := NewGauge(opts)
	if gauge == nil {
		t.Fatal("NewGauge() returned nil")
	}

	// 测试方法不会 panic
	gauge.Set(10.5)
	gauge.Inc()
	gauge.Dec()
	gauge.Add(5.0)
	gauge.Sub(2.0)
}

func TestNewHistogram(t *testing.T) {
	opts := HistogramOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "histogram",
		Help:      "test histogram",
		Buckets:   DefaultBuckets,
	}

	histogram := NewHistogram(opts)
	if histogram == nil {
		t.Fatal("NewHistogram() returned nil")
	}

	// 测试方法不会 panic
	histogram.Observe(10.5)
	histogram.Observe(100.0)
}

func TestNewSummary(t *testing.T) {
	opts := SummaryOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "summary",
		Help:      "test summary",
	}

	summary := NewSummary(opts)
	if summary == nil {
		t.Fatal("NewSummary() returned nil")
	}

	// 测试方法不会 panic
	summary.Observe(10.5)
	summary.Observe(100.0)
}

func TestNewCounterVec(t *testing.T) {
	opts := CounterOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "counter_vec",
		Help:      "test counter vec",
	}

	vec := NewCounterVec(opts, []string{"label1", "label2"})
	if vec == nil {
		t.Fatal("NewCounterVec() returned nil")
	}

	// 测试 WithLabelValues
	counter := vec.WithLabelValues("value1", "value2")
	if counter == nil {
		t.Fatal("WithLabelValues() returned nil")
	}

	counter.Inc()
	counter.Add(1.0)
}

func TestNewGaugeVec(t *testing.T) {
	opts := GaugeOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "gauge_vec",
		Help:      "test gauge vec",
	}

	vec := NewGaugeVec(opts, []string{"label1"})
	if vec == nil {
		t.Fatal("NewGaugeVec() returned nil")
	}

	gauge := vec.WithLabelValues("value1")
	if gauge == nil {
		t.Fatal("WithLabelValues() returned nil")
	}

	gauge.Set(10.0)
}

func TestNewHistogramVec(t *testing.T) {
	opts := HistogramOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "histogram_vec",
		Help:      "test histogram vec",
		Buckets:   DefaultBuckets,
	}

	vec := NewHistogramVec(opts, []string{"label1"})
	if vec == nil {
		t.Fatal("NewHistogramVec() returned nil")
	}

	histogram := vec.WithLabelValues("value1")
	if histogram == nil {
		t.Fatal("WithLabelValues() returned nil")
	}

	histogram.Observe(10.0)
}

func TestNewSummaryVec(t *testing.T) {
	opts := SummaryOpts{
		Namespace: "test",
		Subsystem: "test",
		Name:      "summary_vec",
		Help:      "test summary vec",
	}

	vec := NewSummaryVec(opts, []string{"label1"})
	if vec == nil {
		t.Fatal("NewSummaryVec() returned nil")
	}

	summary := vec.WithLabelValues("value1")
	if summary == nil {
		t.Fatal("WithLabelValues() returned nil")
	}

	summary.Observe(10.0)
}

func TestNewTimer(t *testing.T) {
	histogram := NewHistogram(HistogramOpts{
		Namespace: "test",
		Name:      "timer",
		Help:      "test timer",
	})

	timer := NewTimer(histogram)
	if timer == nil {
		t.Fatal("NewTimer() returned nil")
	}

	// 测试 ObserveDuration 不会 panic
	timer.ObserveDuration()

	// 测试 nil histogram
	timer2 := NewTimer(nil)
	if timer2 == nil {
		t.Fatal("NewTimer() with nil should return non-nil")
	}
	timer2.ObserveDuration() // 应该不会 panic
}

func TestSetProvider(t *testing.T) {
	original := GetProvider()

	// 设置新的 provider
	newProvider := &noopProvider{}
	SetProvider(newProvider)

	if GetProvider() != newProvider {
		t.Fatal("Provider not set correctly")
	}

	// 恢复
	SetProvider(original)
}

func TestGetProvider(t *testing.T) {
	provider := GetProvider()
	if provider == nil {
		t.Fatal("GetProvider() returned nil")
	}
}
