/**
 * @Author: dawn
 * @Desc: Tracing 抽象层定义
 */

package tracing

import (
	"context"
)

// ============================================================================
// Span 状态和类型
// ============================================================================

// SpanKind Span 类型
type SpanKind int

const (
	SpanKindUnspecified SpanKind = iota
	SpanKindInternal             // 内部 Span
	SpanKindServer               // 服务端 Span
	SpanKindClient               // 客户端 Span
	SpanKindProducer             // 生产者 Span
	SpanKindConsumer             // 消费者 Span
)

// SpanStatus Span 状态
type SpanStatus int

const (
	SpanStatusUnset SpanStatus = iota
	SpanStatusOK
	SpanStatusError
)

// ============================================================================
// 核心接口
// ============================================================================

// ITracer 追踪器接口
type ITracer interface {
	// Start 开始一个新的 Span
	Start(ctx context.Context, spanName string, opts ...SpanStartOption) (context.Context, Span)
}

// Span 表示一个追踪单元
type Span interface {
	// End 结束 Span
	End(opts ...SpanEndOption)

	// SetName 设置 Span 名称
	SetName(name string)

	// SetStatus 设置 Span 状态
	SetStatus(status SpanStatus, description string)

	// SetAttributes 设置属性
	SetAttributes(kv ...Attribute)

	// AddEvent 添加事件
	AddEvent(name string, attrs ...Attribute)

	// RecordError 记录错误
	RecordError(err error, opts ...EventOption)

	// SpanContext 获取 Span 上下文
	SpanContext() SpanContext

	// IsRecording 是否正在记录
	IsRecording() bool
}

// SpanContext Span 上下文
type SpanContext interface {
	// TraceID 获取追踪 ID
	TraceID() string
	// SpanID 获取 Span ID
	SpanID() string
	// IsValid 是否有效
	IsValid() bool
	// IsSampled 是否采样
	IsSampled() bool
}

// ============================================================================
// 属性定义
// ============================================================================

// Attribute 键值对属性
type Attribute struct {
	Key   string
	Value any
}

// String 创建字符串属性
func String(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int 创建整数属性
func Int(key string, value int) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int64 创建 int64 属性
func Int64(key string, value int64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Float64 创建浮点数属性
func Float64(key string, value float64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Bool 创建布尔属性
func Bool(key string, value bool) Attribute {
	return Attribute{Key: key, Value: value}
}

// StringSlice 创建字符串切片属性
func StringSlice(key string, value []string) Attribute {
	return Attribute{Key: key, Value: value}
}

// ============================================================================
// 选项定义
// ============================================================================

// SpanStartOption Span 开始选项
type SpanStartOption interface {
	ApplySpanStart(*SpanStartConfig)
}

// SpanEndOption Span 结束选项
type SpanEndOption interface {
	ApplySpanEnd(*SpanEndConfig)
}

// EventOption 事件选项
type EventOption interface {
	ApplyEvent(*EventConfig)
}

// SpanStartConfig Span 开始配置
type SpanStartConfig struct {
	Kind       SpanKind
	Attributes []Attribute
}

// SpanEndConfig Span 结束配置
type SpanEndConfig struct{}

// EventConfig 事件配置
type EventConfig struct {
	Attributes []Attribute
}

// WithSpanKind 设置 Span 类型
func WithSpanKind(kind SpanKind) SpanStartOption {
	return SpanKindOption(kind)
}

// SpanKindOption Span 类型选项
type SpanKindOption SpanKind

// ApplySpanStart 应用到配置
func (o SpanKindOption) ApplySpanStart(cfg *SpanStartConfig) {
	cfg.Kind = SpanKind(o)
}

// WithAttributes 设置属性
func WithAttributes(attrs ...Attribute) SpanStartOption {
	return AttributesOption(attrs)
}

// AttributesOption 属性选项
type AttributesOption []Attribute

// ApplySpanStart 应用到配置
func (o AttributesOption) ApplySpanStart(cfg *SpanStartConfig) {
	cfg.Attributes = append(cfg.Attributes, o...)
}

// ============================================================================
// Provider 接口
// ============================================================================

// TracerProvider 追踪器提供者接口
type TracerProvider interface {
	// Tracer 获取追踪器
	Tracer(name string, opts ...TracerOption) ITracer
	// Shutdown 关闭提供者
	Shutdown(ctx context.Context) error
}

// TracerOption 追踪器选项
type TracerOption interface {
	ApplyTracer(*TracerConfig)
}

// TracerConfig 追踪器配置
type TracerConfig struct {
	Version string
}

// WithVersion 设置版本
func WithVersion(version string) TracerOption {
	return VersionOption(version)
}

// VersionOption 版本选项
type VersionOption string

// ApplyTracer 应用到配置
func (o VersionOption) ApplyTracer(cfg *TracerConfig) {
	cfg.Version = string(o)
}

// ============================================================================
// 语义约定属性名
// ============================================================================

const (
	// 通用属性
	AttrServiceName      = "service.name"
	AttrServiceVersion   = "service.version"
	AttrServiceInstance  = "service.instance.id"

	// 网络属性
	AttrNetPeerName      = "net.peer.name"
	AttrNetPeerPort      = "net.peer.port"
	AttrNetHostName      = "net.host.name"
	AttrNetHostPort      = "net.host.port"
	AttrNetTransport     = "net.transport"

	// RPC 属性
	AttrRPCSystem        = "rpc.system"
	AttrRPCService       = "rpc.service"
	AttrRPCMethod        = "rpc.method"

	// HTTP 属性
	AttrHTTPMethod       = "http.method"
	AttrHTTPURL          = "http.url"
	AttrHTTPTarget       = "http.target"
	AttrHTTPStatusCode   = "http.status_code"
	AttrHTTPRequestSize  = "http.request_content_length"
	AttrHTTPResponseSize = "http.response_content_length"

	// 消息属性
	AttrMessageType      = "message.type"
	AttrMessageID        = "message.id"
	AttrMessageSize      = "message.uncompressed_size"
)
