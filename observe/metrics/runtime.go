/**
 * @Author: dawn
 * @Desc: 运行时性能指标收集，包括内存、goroutine、GC等系统指标
 */

package metrics

import (
	"runtime"
	"sync"
	"time"
)

// ============================================================================
// 运行时指标定义
// ============================================================================

const (
	SubsystemRuntime = "runtime"
)

var (
	// RuntimeGoroutines goroutine 数量
	RuntimeGoroutines Gauge

	// RuntimeThreads 线程数量
	RuntimeThreads Gauge

	// RuntimeMemAllocBytes 当前分配的内存字节数
	RuntimeMemAllocBytes Gauge

	// RuntimeMemTotalAllocBytes 累计分配的内存字节数
	RuntimeMemTotalAllocBytes Counter

	// RuntimeMemSysBytes 从系统获取的内存字节数
	RuntimeMemSysBytes Gauge

	// RuntimeMemHeapAllocBytes 堆分配的内存字节数
	RuntimeMemHeapAllocBytes Gauge

	// RuntimeMemHeapSysBytes 堆从系统获取的内存字节数
	RuntimeMemHeapSysBytes Gauge

	// RuntimeMemHeapIdleBytes 堆空闲内存字节数
	RuntimeMemHeapIdleBytes Gauge

	// RuntimeMemHeapInuseBytes 堆正在使用的内存字节数
	RuntimeMemHeapInuseBytes Gauge

	// RuntimeMemHeapObjects 堆对象数量
	RuntimeMemHeapObjects Gauge

	// RuntimeMemStackInuseBytes 栈正在使用的内存字节数
	RuntimeMemStackInuseBytes Gauge

	// RuntimeGCPauseTotal GC 暂停总时间（纳秒）
	RuntimeGCPauseTotal Counter

	// RuntimeGCPauseLast 上次 GC 暂停时间（纳秒）
	RuntimeGCPauseLast Gauge

	// RuntimeGCCount GC 执行次数
	RuntimeGCCount Counter

	// RuntimeGCCPUFraction GC 使用的 CPU 比例
	RuntimeGCCPUFraction Gauge
)

var (
	runtimeCollectorOnce sync.Once
	runtimeCollector     *RuntimeCollector
)

// ============================================================================
// 运行时指标收集器
// ============================================================================

// RuntimeCollector 运行时指标收集器
type RuntimeCollector struct {
	interval time.Duration
	stopCh   chan struct{}
	stopped  bool
	mu       sync.Mutex

	// 上一次的统计值（用于计算增量）
	lastTotalAlloc uint64
	lastGCPause    uint64
	lastNumGC      uint32
}

// RuntimeCollectorOption 收集器选项
type RuntimeCollectorOption func(*RuntimeCollector)

// WithCollectInterval 设置收集间隔
func WithCollectInterval(interval time.Duration) RuntimeCollectorOption {
	return func(c *RuntimeCollector) {
		c.interval = interval
	}
}

// NewRuntimeCollector 创建运行时指标收集器
func NewRuntimeCollector(opts ...RuntimeCollectorOption) *RuntimeCollector {
	c := &RuntimeCollector{
		interval: 15 * time.Second,
		stopCh:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// InitRuntimeMetrics 初始化运行时指标
func InitRuntimeMetrics() {
	RuntimeGoroutines = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "goroutines",
		Help:      "Number of goroutines",
	})

	RuntimeThreads = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "threads",
		Help:      "Number of OS threads",
	})

	RuntimeMemAllocBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_alloc_bytes",
		Help:      "Current allocated memory in bytes",
	})

	RuntimeMemTotalAllocBytes = NewCounter(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_total_alloc_bytes",
		Help:      "Total allocated memory in bytes",
	})

	RuntimeMemSysBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_sys_bytes",
		Help:      "Memory obtained from system in bytes",
	})

	RuntimeMemHeapAllocBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_heap_alloc_bytes",
		Help:      "Heap allocated memory in bytes",
	})

	RuntimeMemHeapSysBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_heap_sys_bytes",
		Help:      "Heap memory obtained from system in bytes",
	})

	RuntimeMemHeapIdleBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_heap_idle_bytes",
		Help:      "Heap idle memory in bytes",
	})

	RuntimeMemHeapInuseBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_heap_inuse_bytes",
		Help:      "Heap in-use memory in bytes",
	})

	RuntimeMemHeapObjects = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_heap_objects",
		Help:      "Number of heap objects",
	})

	RuntimeMemStackInuseBytes = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "mem_stack_inuse_bytes",
		Help:      "Stack in-use memory in bytes",
	})

	RuntimeGCPauseTotal = NewCounter(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "gc_pause_total_nanoseconds",
		Help:      "Total GC pause time in nanoseconds",
	})

	RuntimeGCPauseLast = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "gc_pause_last_nanoseconds",
		Help:      "Last GC pause time in nanoseconds",
	})

	RuntimeGCCount = NewCounter(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "gc_count_total",
		Help:      "Total number of GC cycles",
	})

	RuntimeGCCPUFraction = NewGauge(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemRuntime,
		Name:      "gc_cpu_fraction",
		Help:      "Fraction of CPU time used by GC",
	})
}

// Start 启动收集
func (c *RuntimeCollector) Start() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.mu.Unlock()

	go c.collect()
}

// Stop 停止收集
func (c *RuntimeCollector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.stopped {
		c.stopped = true
		close(c.stopCh)
	}
}

// collect 收集指标
func (c *RuntimeCollector) collect() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// 立即收集一次
	c.doCollect()

	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.doCollect()
		}
	}
}

// doCollect 执行收集
func (c *RuntimeCollector) doCollect() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Goroutines 和线程
	RuntimeGoroutines.Set(float64(runtime.NumGoroutine()))
	RuntimeThreads.Set(float64(runtime.GOMAXPROCS(0)))

	// 内存指标
	RuntimeMemAllocBytes.Set(float64(memStats.Alloc))
	RuntimeMemSysBytes.Set(float64(memStats.Sys))
	RuntimeMemHeapAllocBytes.Set(float64(memStats.HeapAlloc))
	RuntimeMemHeapSysBytes.Set(float64(memStats.HeapSys))
	RuntimeMemHeapIdleBytes.Set(float64(memStats.HeapIdle))
	RuntimeMemHeapInuseBytes.Set(float64(memStats.HeapInuse))
	RuntimeMemHeapObjects.Set(float64(memStats.HeapObjects))
	RuntimeMemStackInuseBytes.Set(float64(memStats.StackInuse))

	// 计算增量
	if memStats.TotalAlloc > c.lastTotalAlloc {
		RuntimeMemTotalAllocBytes.Add(float64(memStats.TotalAlloc - c.lastTotalAlloc))
	}
	c.lastTotalAlloc = memStats.TotalAlloc

	// GC 指标
	if memStats.NumGC > c.lastNumGC {
		RuntimeGCCount.Add(float64(memStats.NumGC - c.lastNumGC))
	}
	c.lastNumGC = memStats.NumGC

	RuntimeGCCPUFraction.Set(memStats.GCCPUFraction)

	// GC 暂停时间
	if memStats.PauseTotalNs > c.lastGCPause {
		RuntimeGCPauseTotal.Add(float64(memStats.PauseTotalNs - c.lastGCPause))
	}
	c.lastGCPause = memStats.PauseTotalNs

	// 最近一次 GC 暂停时间
	if memStats.NumGC > 0 {
		idx := (memStats.NumGC + 255) % 256
		RuntimeGCPauseLast.Set(float64(memStats.PauseNs[idx]))
	}
}

// StartRuntimeCollector 启动全局运行时收集器
func StartRuntimeCollector(opts ...RuntimeCollectorOption) {
	runtimeCollectorOnce.Do(func() {
		InitRuntimeMetrics()
		runtimeCollector = NewRuntimeCollector(opts...)
		runtimeCollector.Start()
	})
}

// StopRuntimeCollector 停止全局运行时收集器
func StopRuntimeCollector() {
	if runtimeCollector != nil {
		runtimeCollector.Stop()
	}
}

// GetRuntimeStats 获取当前运行时统计信息
func GetRuntimeStats() *RuntimeStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &RuntimeStats{
		Goroutines:     runtime.NumGoroutine(),
		NumCPU:         runtime.NumCPU(),
		GOMAXPROCS:     runtime.GOMAXPROCS(0),
		MemAlloc:       memStats.Alloc,
		MemTotalAlloc:  memStats.TotalAlloc,
		MemSys:         memStats.Sys,
		HeapAlloc:      memStats.HeapAlloc,
		HeapSys:        memStats.HeapSys,
		HeapIdle:       memStats.HeapIdle,
		HeapInuse:      memStats.HeapInuse,
		HeapObjects:    memStats.HeapObjects,
		StackInuse:     memStats.StackInuse,
		NumGC:          memStats.NumGC,
		GCCPUFraction:  memStats.GCCPUFraction,
		LastGCPauseNs:  memStats.PauseNs[(memStats.NumGC+255)%256],
	}
}

// RuntimeStats 运行时统计信息
type RuntimeStats struct {
	Goroutines     int
	NumCPU         int
	GOMAXPROCS     int
	MemAlloc       uint64
	MemTotalAlloc  uint64
	MemSys         uint64
	HeapAlloc      uint64
	HeapSys        uint64
	HeapIdle       uint64
	HeapInuse      uint64
	HeapObjects    uint64
	StackInuse     uint64
	NumGC          uint32
	GCCPUFraction  float64
	LastGCPauseNs  uint64
}
