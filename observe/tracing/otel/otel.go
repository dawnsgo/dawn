/**
 * @Author: dawn
 * @Desc: OpenTelemetry 追踪器适配器
 */

package otel

import (
	"context"

	"github.com/dawnsgo/dawn/observe/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Provider OpenTelemetry 追踪器提供者
type Provider struct {
	tp oteltrace.TracerProvider
}

// Option 配置选项
type Option func(*Provider)

// WithTracerProvider 设置 OTel TracerProvider
func WithTracerProvider(tp oteltrace.TracerProvider) Option {
	return func(p *Provider) {
		p.tp = tp
	}
}

// NewProvider 创建 OpenTelemetry 提供者
func NewProvider(opts ...Option) *Provider {
	p := &Provider{
		tp: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Tracer 获取追踪器
func (p *Provider) Tracer(name string, opts ...tracing.TracerOption) tracing.ITracer {
	var cfg tracing.TracerConfig
	for _, opt := range opts {
		opt.ApplyTracer(&cfg)
	}

	var tracerOpts []oteltrace.TracerOption
	if cfg.Version != "" {
		tracerOpts = append(tracerOpts, oteltrace.WithInstrumentationVersion(cfg.Version))
	}

	return &tracer{
		tracer: p.tp.Tracer(name, tracerOpts...),
	}
}

// Shutdown 关闭提供者
func (p *Provider) Shutdown(ctx context.Context) error {
	if sp, ok := p.tp.(interface{ Shutdown(context.Context) error }); ok {
		return sp.Shutdown(ctx)
	}
	return nil
}

// ============================================================================
// Tracer 适配器
// ============================================================================

type tracer struct {
	tracer oteltrace.Tracer
}

func (t *tracer) Start(ctx context.Context, spanName string, opts ...tracing.SpanStartOption) (context.Context, tracing.Span) {
	var cfg tracing.SpanStartConfig
	for _, opt := range opts {
		opt.ApplySpanStart(&cfg)
	}

	var spanOpts []oteltrace.SpanStartOption
	spanOpts = append(spanOpts, oteltrace.WithSpanKind(convertSpanKind(cfg.Kind)))
	if len(cfg.Attributes) > 0 {
		spanOpts = append(spanOpts, oteltrace.WithAttributes(convertAttributes(cfg.Attributes)...))
	}

	ctx, otelSpan := t.tracer.Start(ctx, spanName, spanOpts...)
	return ctx, &span{span: otelSpan}
}

// ============================================================================
// Span 适配器
// ============================================================================

type span struct {
	span oteltrace.Span
}

func (s *span) End(opts ...tracing.SpanEndOption) {
	s.span.End()
}

func (s *span) SetName(name string) {
	s.span.SetName(name)
}

func (s *span) SetStatus(status tracing.SpanStatus, description string) {
	s.span.SetStatus(convertSpanStatus(status), description)
}

func (s *span) SetAttributes(kv ...tracing.Attribute) {
	s.span.SetAttributes(convertAttributes(kv)...)
}

func (s *span) AddEvent(name string, attrs ...tracing.Attribute) {
	s.span.AddEvent(name, oteltrace.WithAttributes(convertAttributes(attrs)...))
}

func (s *span) RecordError(err error, opts ...tracing.EventOption) {
	s.span.RecordError(err)
}

func (s *span) SpanContext() tracing.SpanContext {
	return &spanContext{ctx: s.span.SpanContext()}
}

func (s *span) IsRecording() bool {
	return s.span.IsRecording()
}

// ============================================================================
// SpanContext 适配器
// ============================================================================

type spanContext struct {
	ctx oteltrace.SpanContext
}

func (c *spanContext) TraceID() string {
	return c.ctx.TraceID().String()
}

func (c *spanContext) SpanID() string {
	return c.ctx.SpanID().String()
}

func (c *spanContext) IsValid() bool {
	return c.ctx.IsValid()
}

func (c *spanContext) IsSampled() bool {
	return c.ctx.IsSampled()
}

// ============================================================================
// 转换函数
// ============================================================================

func convertSpanKind(kind tracing.SpanKind) oteltrace.SpanKind {
	switch kind {
	case tracing.SpanKindInternal:
		return oteltrace.SpanKindInternal
	case tracing.SpanKindServer:
		return oteltrace.SpanKindServer
	case tracing.SpanKindClient:
		return oteltrace.SpanKindClient
	case tracing.SpanKindProducer:
		return oteltrace.SpanKindProducer
	case tracing.SpanKindConsumer:
		return oteltrace.SpanKindConsumer
	default:
		return oteltrace.SpanKindUnspecified
	}
}

func convertSpanStatus(status tracing.SpanStatus) codes.Code {
	switch status {
	case tracing.SpanStatusOK:
		return codes.Ok
	case tracing.SpanStatusError:
		return codes.Error
	default:
		return codes.Unset
	}
}

func convertAttributes(attrs []tracing.Attribute) []attribute.KeyValue {
	result := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		result = append(result, convertAttribute(attr))
	}
	return result
}

func convertAttribute(attr tracing.Attribute) attribute.KeyValue {
	switch v := attr.Value.(type) {
	case string:
		return attribute.String(attr.Key, v)
	case int:
		return attribute.Int(attr.Key, v)
	case int64:
		return attribute.Int64(attr.Key, v)
	case float64:
		return attribute.Float64(attr.Key, v)
	case bool:
		return attribute.Bool(attr.Key, v)
	case []string:
		return attribute.StringSlice(attr.Key, v)
	default:
		return attribute.String(attr.Key, "")
	}
}
