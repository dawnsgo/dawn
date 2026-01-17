/**
 * @Author: dawn
 * @Desc: Tracing 全局注册表
 */

package tracing

import (
	"context"
	"sync"
)

var (
	globalProvider TracerProvider = &noopTracerProvider{}
	providerMu     sync.RWMutex
)

// SetTracerProvider 设置全局追踪器提供者
func SetTracerProvider(p TracerProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	globalProvider = p
}

// GetTracerProvider 获取全局追踪器提供者
func GetTracerProvider() TracerProvider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return globalProvider
}

// Tracer 获取追踪器
func Tracer(name string, opts ...TracerOption) ITracer {
	return GetTracerProvider().Tracer(name, opts...)
}

// Start 开始一个新的 Span（便捷函数）
func Start(ctx context.Context, tracerName, spanName string, opts ...SpanStartOption) (context.Context, Span) {
	return Tracer(tracerName).Start(ctx, spanName, opts...)
}

// ============================================================================
// 上下文操作
// ============================================================================

type spanContextKey struct{}

// SpanFromContext 从上下文获取 Span
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanContextKey{}).(Span); ok {
		return span
	}
	return &noopSpan{}
}

// ContextWithSpan 将 Span 放入上下文
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

// ============================================================================
// Noop 实现
// ============================================================================

type noopTracerProvider struct{}

func (p *noopTracerProvider) Tracer(name string, opts ...TracerOption) ITracer {
	return &noopTracer{}
}

func (p *noopTracerProvider) Shutdown(ctx context.Context) error {
	return nil
}

type noopTracer struct{}

func (t *noopTracer) Start(ctx context.Context, spanName string, opts ...SpanStartOption) (context.Context, Span) {
	span := &noopSpan{}
	return ContextWithSpan(ctx, span), span
}

type noopSpan struct{}

func (s *noopSpan) End(opts ...SpanEndOption)                        {}
func (s *noopSpan) SetName(name string)                              {}
func (s *noopSpan) SetStatus(status SpanStatus, description string)  {}
func (s *noopSpan) SetAttributes(kv ...Attribute)                    {}
func (s *noopSpan) AddEvent(name string, attrs ...Attribute)         {}
func (s *noopSpan) RecordError(err error, opts ...EventOption)       {}
func (s *noopSpan) SpanContext() SpanContext                         { return &noopSpanContext{} }
func (s *noopSpan) IsRecording() bool                                { return false }

type noopSpanContext struct{}

func (c *noopSpanContext) TraceID() string { return "" }
func (c *noopSpanContext) SpanID() string  { return "" }
func (c *noopSpanContext) IsValid() bool   { return false }
func (c *noopSpanContext) IsSampled() bool { return false }
