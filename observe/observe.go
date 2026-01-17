/**
 * @Author: dawn
 * @Desc: 可观测性模块汇总
 */

package observe

import (
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
