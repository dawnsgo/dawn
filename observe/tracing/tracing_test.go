/**
 * @Author: dawn
 * @Desc: Tracing 单元测试
 */

package tracing

import (
	"context"
	"testing"
)

func TestSpanKind(t *testing.T) {
	// 测试各种 SpanKind 常量存在
	_ = SpanKindUnspecified
	_ = SpanKindInternal
	_ = SpanKindServer
	_ = SpanKindClient
	_ = SpanKindProducer
	_ = SpanKindConsumer
}

func TestNewTracer(t *testing.T) {
	tracer := Tracer("test")
	if tracer == nil {
		t.Fatal("Tracer() returned nil")
	}
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	tracer := Tracer("test")

	newCtx, span := tracer.Start(ctx, "test-span")
	if newCtx == nil {
		t.Fatal("Start() returned nil context")
	}
	if span == nil {
		t.Fatal("Start() returned nil span")
	}

	// 测试 Span 方法不会 panic
	span.SetName("new-name")
	span.SetStatus(SpanStatusOK, "ok")
	span.SetAttributes(String("key", "value"))
	span.AddEvent("event", String("attr", "value"))
	span.RecordError(nil)
	span.End()

	// 测试 SpanContext
	spanCtx := span.SpanContext()
	if spanCtx == nil {
		t.Fatal("SpanContext() returned nil")
	}

	// 测试 IsRecording
	_ = span.IsRecording()
}

func TestStartWithOptions(t *testing.T) {
	ctx := context.Background()
	tracer := Tracer("test")

	// 测试带选项的 Start
	newCtx, span := tracer.Start(ctx, "test-span",
		WithSpanKind(SpanKindServer),
		WithAttributes(String("key1", "value1"), Int("key2", 42)),
	)
	if newCtx == nil {
		t.Fatal("Start() returned nil context")
	}
	if span == nil {
		t.Fatal("Start() returned nil span")
	}

	span.End()
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()

	// 测试从空上下文获取
	span := SpanFromContext(ctx)
	if span == nil {
		t.Fatal("SpanFromContext() should return noop span, not nil")
	}

	// 测试从带 Span 的上下文获取
	tracer := Tracer("test")
	newCtx, originalSpan := tracer.Start(ctx, "test-span")
	span = SpanFromContext(newCtx)
	if span == nil {
		t.Fatal("SpanFromContext() returned nil")
	}
	originalSpan.End()
}

func TestContextWithSpan(t *testing.T) {
	ctx := context.Background()
	tracer := Tracer("test")
	_, span := tracer.Start(ctx, "test-span")

	newCtx := ContextWithSpan(ctx, span)
	if newCtx == nil {
		t.Fatal("ContextWithSpan() returned nil")
	}

	// 验证可以从新上下文获取 Span
	retrievedSpan := SpanFromContext(newCtx)
	if retrievedSpan == nil {
		t.Fatal("SpanFromContext() returned nil after ContextWithSpan()")
	}

	span.End()
}

func TestSetTracerProvider(t *testing.T) {
	original := GetTracerProvider()

	// 设置新的 provider
	newProvider := &noopTracerProvider{}
	SetTracerProvider(newProvider)

	if GetTracerProvider() != newProvider {
		t.Fatal("TracerProvider not set correctly")
	}

	// 恢复
	SetTracerProvider(original)
}

func TestGetTracerProvider(t *testing.T) {
	provider := GetTracerProvider()
	if provider == nil {
		t.Fatal("GetTracerProvider() returned nil")
	}
}

func TestAttributeHelpers(t *testing.T) {
	// 测试各种属性创建函数
	attrs := []Attribute{
		String("str", "value"),
		Int("int", 42),
		Int64("int64", 123456789),
		Float64("float64", 3.14),
		Bool("bool", true),
		StringSlice("slice", []string{"a", "b", "c"}),
	}

	for _, attr := range attrs {
		if attr.Key == "" {
			t.Fatal("Attribute Key should not be empty")
		}
		if attr.Value == nil {
			t.Fatal("Attribute Value should not be nil")
		}
	}
}

func TestSpanStatus(t *testing.T) {
	// 测试 SpanStatus 常量
	statuses := []SpanStatus{
		SpanStatusUnset,
		SpanStatusOK,
		SpanStatusError,
	}

	for _, status := range statuses {
		_ = status // 确保常量存在
	}
}

func TestSpanKindConstants(t *testing.T) {
	// 测试 SpanKind 常量
	kinds := []SpanKind{
		SpanKindUnspecified,
		SpanKindInternal,
		SpanKindServer,
		SpanKindClient,
		SpanKindProducer,
		SpanKindConsumer,
	}

	for _, kind := range kinds {
		_ = kind // 确保常量存在
	}
}

func TestNoopSpan(t *testing.T) {
	span := &noopSpan{}

	// 测试所有方法不会 panic
	span.End()
	span.SetName("test")
	span.SetStatus(SpanStatusOK, "ok")
	span.SetAttributes(String("key", "value"))
	span.AddEvent("event")
	span.RecordError(nil)

	spanCtx := span.SpanContext()
	if spanCtx == nil {
		t.Fatal("SpanContext() should return non-nil")
	}

	if span.IsRecording() {
		t.Fatal("noopSpan should not be recording")
	}
}

func TestNoopSpanContext(t *testing.T) {
	ctx := &noopSpanContext{}

	if ctx.TraceID() != "" {
		t.Fatal("noopSpanContext TraceID should be empty")
	}
	if ctx.SpanID() != "" {
		t.Fatal("noopSpanContext SpanID should be empty")
	}
	if ctx.IsValid() {
		t.Fatal("noopSpanContext should not be valid")
	}
	if ctx.IsSampled() {
		t.Fatal("noopSpanContext should not be sampled")
	}
}
