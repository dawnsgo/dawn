/**
 * @Author: dawn
 * @Desc: 可观测性模块汇总，提供 Metrics 和 Tracing 的统一入口
 */

package observe

import (
	"context"

	"github.com/dawnsgo/dawn/observe/metrics"
	"github.com/dawnsgo/dawn/observe/tracing"
)

// ============================================================================
// Metrics 便捷导出
// ============================================================================

// SetMetricsProvider 设置指标提供者
var SetMetricsProvider = metrics.SetProvider

// GetMetricsProvider 获取指标提供者
var GetMetricsProvider = metrics.GetProvider

// InitBuiltinMetrics 初始化内置指标
var InitBuiltinMetrics = metrics.InitBuiltinMetrics

// InitRuntimeMetrics 初始化运行时指标
var InitRuntimeMetrics = metrics.InitRuntimeMetrics

// InitPerformanceMetrics 初始化性能指标
var InitPerformanceMetrics = metrics.InitPerformanceMetrics

// StartRuntimeCollector 启动运行时指标收集器
var StartRuntimeCollector = metrics.StartRuntimeCollector

// StopRuntimeCollector 停止运行时指标收集器
var StopRuntimeCollector = metrics.StopRuntimeCollector

// GetRuntimeStats 获取运行时统计信息
var GetRuntimeStats = metrics.GetRuntimeStats

// GetPerformanceReport 获取性能报告
var GetPerformanceReport = metrics.GetPerformanceReport

// ============================================================================
// 性能追踪便捷导出
// ============================================================================

// StartTracker 创建性能追踪器
var StartTracker = metrics.StartTracker

// TrackOperation 追踪操作
var TrackOperation = metrics.TrackOperation

// MeasureFunc 测量函数执行时间
var MeasureFunc = metrics.MeasureFunc

// ============================================================================
// Tracing 便捷导出
// ============================================================================

// SetTracerProvider 设置追踪器提供者
var SetTracerProvider = tracing.SetTracerProvider

// GetTracerProvider 获取追踪器提供者
var GetTracerProvider = tracing.GetTracerProvider

// Tracer 获取追踪器
var Tracer = tracing.Tracer

// SpanFromContext 从上下文获取 Span
var SpanFromContext = tracing.SpanFromContext

// ContextWithSpan 将 Span 放入上下文
var ContextWithSpan = tracing.ContextWithSpan

// StartSpan 开始一个新的 Span
var StartSpan = tracing.StartSpan

// WrapSpan 包装函数执行并自动追踪
var WrapSpan = tracing.WrapSpan

// NewSpanBuilder 创建 Span 构建器
var NewSpanBuilder = tracing.NewSpanBuilder

// ExtractTraceInfo 提取追踪信息
var ExtractTraceInfo = tracing.ExtractTraceInfo

// ============================================================================
// 统一初始化
// ============================================================================

// InitAllMetrics 初始化所有指标
func InitAllMetrics() {
	metrics.InitBuiltinMetrics()
	metrics.InitRuntimeMetrics()
	metrics.InitPerformanceMetrics()
}

// StartAllCollectors 启动所有收集器
func StartAllCollectors() {
	metrics.StartRuntimeCollector()
}

// StopAllCollectors 停止所有收集器
func StopAllCollectors() {
	metrics.StopRuntimeCollector()
}

// Shutdown 优雅关闭可观测性模块
func Shutdown(ctx context.Context) error {
	StopAllCollectors()

	if tp := tracing.GetTracerProvider(); tp != nil {
		return tp.Shutdown(ctx)
	}
	return nil
}
