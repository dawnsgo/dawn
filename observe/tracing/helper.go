/**
 * @Author: dawn
 * @Desc: 追踪辅助函数，提供便捷的分布式追踪功能
 */

package tracing

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ============================================================================
// 追踪辅助函数
// ============================================================================

// StartSpan 开始一个新的 Span（便捷函数）
func StartSpan(ctx context.Context, name string, opts ...SpanStartOption) (context.Context, Span) {
	return Tracer("dawn").Start(ctx, name, opts...)
}

// StartSpanWithTracer 使用指定追踪器开始一个新的 Span
func StartSpanWithTracer(ctx context.Context, tracerName, spanName string, opts ...SpanStartOption) (context.Context, Span) {
	return Tracer(tracerName).Start(ctx, spanName, opts...)
}

// WrapSpan 包装函数执行并自动追踪
func WrapSpan(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(SpanStatusError, err.Error())
	} else {
		span.SetStatus(SpanStatusOK, "")
	}

	return err
}

// WrapSpanWithResult 包装带返回值的函数执行并自动追踪
func WrapSpanWithResult[T any](ctx context.Context, name string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := StartSpan(ctx, name)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(SpanStatusError, err.Error())
	} else {
		span.SetStatus(SpanStatusOK, "")
	}

	return result, err
}

// ============================================================================
// Span 链式构建器
// ============================================================================

// SpanBuilder Span 构建器
type SpanBuilder struct {
	ctx        context.Context
	tracerName string
	spanName   string
	kind       SpanKind
	attrs      []Attribute
}

// NewSpanBuilder 创建 Span 构建器
func NewSpanBuilder(ctx context.Context, spanName string) *SpanBuilder {
	return &SpanBuilder{
		ctx:        ctx,
		tracerName: "dawn",
		spanName:   spanName,
		kind:       SpanKindInternal,
	}
}

// WithTracerName 设置追踪器名称
func (b *SpanBuilder) WithTracerName(name string) *SpanBuilder {
	b.tracerName = name
	return b
}

// WithKind 设置 Span 类型
func (b *SpanBuilder) WithKind(kind SpanKind) *SpanBuilder {
	b.kind = kind
	return b
}

// WithAttribute 添加属性
func (b *SpanBuilder) WithAttribute(key string, value any) *SpanBuilder {
	b.attrs = append(b.attrs, Attribute{Key: key, Value: value})
	return b
}

// WithAttributes 添加多个属性
func (b *SpanBuilder) WithAttributes(attrs ...Attribute) *SpanBuilder {
	b.attrs = append(b.attrs, attrs...)
	return b
}

// Start 开始 Span
func (b *SpanBuilder) Start() (context.Context, Span) {
	opts := []SpanStartOption{
		WithSpanKind(b.kind),
	}
	if len(b.attrs) > 0 {
		opts = append(opts, WithAttributes(b.attrs...))
	}

	return Tracer(b.tracerName).Start(b.ctx, b.spanName, opts...)
}

// ============================================================================
// 自动追踪装饰器
// ============================================================================

// TracedFunc 自动追踪的函数类型
type TracedFunc func(ctx context.Context) error

// Traced 创建自动追踪的函数
func Traced(name string, fn func(ctx context.Context) error) TracedFunc {
	return func(ctx context.Context) error {
		return WrapSpan(ctx, name, fn)
	}
}

// TracedMethod 方法追踪装饰器
func TracedMethod(receiver, method string, fn func(ctx context.Context) error) TracedFunc {
	name := fmt.Sprintf("%s.%s", receiver, method)
	return Traced(name, fn)
}

// ============================================================================
// 调用栈追踪
// ============================================================================

// AddCallerInfo 添加调用者信息到 Span
func AddCallerInfo(span Span, skip int) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return
	}

	fn := runtime.FuncForPC(pc)
	funcName := ""
	if fn != nil {
		funcName = fn.Name()
	}

	span.SetAttributes(
		String("code.filepath", file),
		Int("code.lineno", line),
		String("code.function", funcName),
	)
}

// StartSpanWithCaller 开始 Span 并自动添加调用者信息
func StartSpanWithCaller(ctx context.Context, name string, opts ...SpanStartOption) (context.Context, Span) {
	ctx, span := StartSpan(ctx, name, opts...)
	AddCallerInfo(span, 1)
	return ctx, span
}

// ============================================================================
// 时间追踪
// ============================================================================

// SpanTimer Span 计时器
type SpanTimer struct {
	span      Span
	startTime time.Time
}

// NewSpanTimer 创建 Span 计时器
func NewSpanTimer(span Span) *SpanTimer {
	return &SpanTimer{
		span:      span,
		startTime: time.Now(),
	}
}

// Checkpoint 添加时间检查点事件
func (t *SpanTimer) Checkpoint(name string, attrs ...Attribute) {
	elapsed := time.Since(t.startTime)
	eventAttrs := append(attrs,
		Float64("elapsed_ms", float64(elapsed.Milliseconds())),
	)
	t.span.AddEvent(name, eventAttrs...)
}

// End 结束计时并记录总耗时
func (t *SpanTimer) End() time.Duration {
	elapsed := time.Since(t.startTime)
	t.span.SetAttributes(Float64("duration_ms", float64(elapsed.Milliseconds())))
	t.span.End()
	return elapsed
}

// ============================================================================
// 错误追踪
// ============================================================================

// RecordErrorWithStack 记录错误并添加堆栈信息
func RecordErrorWithStack(span Span, err error) {
	if err == nil || span == nil {
		return
	}

	span.RecordError(err)

	// 获取堆栈
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])

	span.SetAttributes(String("exception.stacktrace", stackTrace))
}

// ============================================================================
// 链路传播
// ============================================================================

// ExtractTraceInfo 从 Span 提取追踪信息
func ExtractTraceInfo(ctx context.Context) (traceID, spanID string, sampled bool) {
	span := SpanFromContext(ctx)
	if span == nil {
		return "", "", false
	}

	sc := span.SpanContext()
	if sc == nil || !sc.IsValid() {
		return "", "", false
	}

	return sc.TraceID(), sc.SpanID(), sc.IsSampled()
}

// TraceInfoToMap 将追踪信息转换为 map（用于跨服务传播）
func TraceInfoToMap(ctx context.Context) map[string]string {
	traceID, spanID, sampled := ExtractTraceInfo(ctx)
	if traceID == "" {
		return nil
	}

	sampledStr := "0"
	if sampled {
		sampledStr = "1"
	}

	return map[string]string{
		"trace-id":      traceID,
		"span-id":       spanID,
		"trace-sampled": sampledStr,
	}
}

// ============================================================================
// 常用 Span 属性设置
// ============================================================================

// SetRPCAttributes 设置 RPC 相关属性
func SetRPCAttributes(span Span, system, service, method string) {
	span.SetAttributes(
		String(AttrRPCSystem, system),
		String(AttrRPCService, service),
		String(AttrRPCMethod, method),
	)
}

// SetHTTPAttributes 设置 HTTP 相关属性
func SetHTTPAttributes(span Span, method, url string, statusCode int) {
	span.SetAttributes(
		String(AttrHTTPMethod, method),
		String(AttrHTTPURL, url),
		Int(AttrHTTPStatusCode, statusCode),
	)
}

// SetNetworkAttributes 设置网络相关属性
func SetNetworkAttributes(span Span, peerName string, peerPort int, transport string) {
	span.SetAttributes(
		String(AttrNetPeerName, peerName),
		Int(AttrNetPeerPort, peerPort),
		String(AttrNetTransport, transport),
	)
}

// SetServiceAttributes 设置服务相关属性
func SetServiceAttributes(span Span, name, version, instanceID string) {
	span.SetAttributes(
		String(AttrServiceName, name),
		String(AttrServiceVersion, version),
		String(AttrServiceInstance, instanceID),
	)
}
