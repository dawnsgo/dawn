/**
 * @Author: dawn
 * @Desc: 性能监控模块，提供组件启动时间、请求延迟等关键路径性能监控
 */

package metrics

import (
	"sync"
	"time"
)

// ============================================================================
// 性能监控子系统
// ============================================================================

const (
	SubsystemPerformance = "performance"
	SubsystemComponent   = "component"
)

// ============================================================================
// 组件启动指标
// ============================================================================

var (
	// ComponentStartupDuration 组件启动耗时（毫秒）
	ComponentStartupDuration HistogramVec

	// ComponentStartupTotal 组件启动次数
	ComponentStartupTotal CounterVec

	// ComponentStartupErrors 组件启动错误次数
	ComponentStartupErrors CounterVec
)

// ============================================================================
// 性能追踪器
// ============================================================================

// PerformanceTracker 性能追踪器
type PerformanceTracker struct {
	name      string
	phase     string
	startTime time.Time
	attrs     map[string]string
}

// StartTracker 创建并启动一个性能追踪器
func StartTracker(name, phase string, attrs ...string) *PerformanceTracker {
	tracker := &PerformanceTracker{
		name:      name,
		phase:     phase,
		startTime: time.Now(),
		attrs:     make(map[string]string),
	}

	for i := 0; i < len(attrs)-1; i += 2 {
		tracker.attrs[attrs[i]] = attrs[i+1]
	}

	return tracker
}

// End 结束追踪并记录指标
func (t *PerformanceTracker) End() time.Duration {
	duration := time.Since(t.startTime)

	if ComponentStartupDuration != nil {
		ComponentStartupDuration.WithLabelValues(t.name, t.phase).Observe(float64(duration.Milliseconds()))
	}

	if ComponentStartupTotal != nil {
		ComponentStartupTotal.WithLabelValues(t.name, t.phase, "success").Inc()
	}

	return duration
}

// EndWithError 结束追踪并记录错误
func (t *PerformanceTracker) EndWithError(err error) time.Duration {
	duration := time.Since(t.startTime)

	if err != nil {
		if ComponentStartupErrors != nil {
			ComponentStartupErrors.WithLabelValues(t.name, t.phase).Inc()
		}
		if ComponentStartupTotal != nil {
			ComponentStartupTotal.WithLabelValues(t.name, t.phase, "error").Inc()
		}
	} else {
		if ComponentStartupTotal != nil {
			ComponentStartupTotal.WithLabelValues(t.name, t.phase, "success").Inc()
		}
	}

	if ComponentStartupDuration != nil {
		ComponentStartupDuration.WithLabelValues(t.name, t.phase).Observe(float64(duration.Milliseconds()))
	}

	return duration
}

// Duration 获取当前耗时
func (t *PerformanceTracker) Duration() time.Duration {
	return time.Since(t.startTime)
}

// ============================================================================
// 操作计时器
// ============================================================================

// OperationTimer 操作计时器
type OperationTimer struct {
	histogram Histogram
	startTime time.Time
}

// NewOperationTimer 创建操作计时器
func NewOperationTimer(h Histogram) *OperationTimer {
	return &OperationTimer{
		histogram: h,
		startTime: time.Now(),
	}
}

// ObserveDuration 观测并记录持续时间（毫秒）
func (t *OperationTimer) ObserveDuration() time.Duration {
	duration := time.Since(t.startTime)
	if t.histogram != nil {
		t.histogram.Observe(float64(duration.Milliseconds()))
	}
	return duration
}

// ObserveDurationSeconds 观测并记录持续时间（秒）
func (t *OperationTimer) ObserveDurationSeconds() time.Duration {
	duration := time.Since(t.startTime)
	if t.histogram != nil {
		t.histogram.Observe(duration.Seconds())
	}
	return duration
}

// Duration 获取当前耗时（不记录）
func (t *OperationTimer) Duration() time.Duration {
	return time.Since(t.startTime)
}

// ============================================================================
// 性能报告
// ============================================================================

// PerformanceReport 性能报告
type PerformanceReport struct {
	mu      sync.RWMutex
	records map[string]*PerformanceRecord
}

// PerformanceRecord 性能记录
type PerformanceRecord struct {
	Name       string
	Phase      string
	Count      int64
	TotalTime  time.Duration
	MinTime    time.Duration
	MaxTime    time.Duration
	LastTime   time.Duration
	ErrorCount int64
}

var (
	globalReport     *PerformanceReport
	globalReportOnce sync.Once
)

// GetPerformanceReport 获取全局性能报告
func GetPerformanceReport() *PerformanceReport {
	globalReportOnce.Do(func() {
		globalReport = &PerformanceReport{
			records: make(map[string]*PerformanceRecord),
		}
	})
	return globalReport
}

// Record 记录性能数据
func (r *PerformanceReport) Record(name, phase string, duration time.Duration, hasError bool) {
	key := name + ":" + phase

	r.mu.Lock()
	defer r.mu.Unlock()

	record, ok := r.records[key]
	if !ok {
		record = &PerformanceRecord{
			Name:    name,
			Phase:   phase,
			MinTime: duration,
			MaxTime: duration,
		}
		r.records[key] = record
	}

	record.Count++
	record.TotalTime += duration
	record.LastTime = duration

	if duration < record.MinTime {
		record.MinTime = duration
	}
	if duration > record.MaxTime {
		record.MaxTime = duration
	}

	if hasError {
		record.ErrorCount++
	}
}

// GetRecord 获取性能记录
func (r *PerformanceReport) GetRecord(name, phase string) *PerformanceRecord {
	key := name + ":" + phase

	r.mu.RLock()
	defer r.mu.RUnlock()

	if record, ok := r.records[key]; ok {
		// 返回副本
		copied := *record
		return &copied
	}
	return nil
}

// GetAllRecords 获取所有性能记录
func (r *PerformanceReport) GetAllRecords() []*PerformanceRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	records := make([]*PerformanceRecord, 0, len(r.records))
	for _, record := range r.records {
		copied := *record
		records = append(records, &copied)
	}
	return records
}

// AvgTime 计算平均耗时
func (r *PerformanceRecord) AvgTime() time.Duration {
	if r.Count == 0 {
		return 0
	}
	return r.TotalTime / time.Duration(r.Count)
}

// ErrorRate 计算错误率
func (r *PerformanceRecord) ErrorRate() float64 {
	if r.Count == 0 {
		return 0
	}
	return float64(r.ErrorCount) / float64(r.Count)
}

// ============================================================================
// 初始化函数
// ============================================================================

// InitPerformanceMetrics 初始化性能指标
func InitPerformanceMetrics() {
	ComponentStartupDuration = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemComponent,
		Name:      "startup_duration_milliseconds",
		Help:      "Component startup duration in milliseconds",
		Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000, 30000, 60000},
	}, []string{"name", "phase"})

	ComponentStartupTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemComponent,
		Name:      "startup_total",
		Help:      "Total number of component startups",
	}, []string{"name", "phase", "status"})

	ComponentStartupErrors = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemComponent,
		Name:      "startup_errors_total",
		Help:      "Total number of component startup errors",
	}, []string{"name", "phase"})
}

// ============================================================================
// 便捷函数
// ============================================================================

// TrackOperation 追踪操作（返回完成函数）
func TrackOperation(h Histogram) func() time.Duration {
	timer := NewOperationTimer(h)
	return timer.ObserveDuration
}

// TrackOperationVec 追踪带标签的操作
func TrackOperationVec(hv HistogramVec, labels ...string) func() time.Duration {
	h := hv.WithLabelValues(labels...)
	timer := NewOperationTimer(h)
	return timer.ObserveDuration
}

// MeasureFunc 测量函数执行时间
func MeasureFunc(name string, fn func() error) (time.Duration, error) {
	tracker := StartTracker(name, "execute")
	err := fn()
	duration := tracker.EndWithError(err)

	// 同时记录到性能报告
	GetPerformanceReport().Record(name, "execute", duration, err != nil)

	return duration, err
}
