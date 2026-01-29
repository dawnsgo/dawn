/**
 * @Author: dawn
 * @Desc: OpenTelemetry OTLP 导出配置
 */

package otel

import (
	"context"
	"time"

	"github.com/dawnsgo/dawn/etc"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/observe/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// ============================================================================
// 配置结构
// ============================================================================

// Config OTLP 导出配置
type Config struct {
	// Enabled 是否启用追踪
	Enabled bool `json:"enabled"`
	// ServiceName 服务名称
	ServiceName string `json:"serviceName"`
	// ServiceVersion 服务版本
	ServiceVersion string `json:"serviceVersion"`
	// Environment 运行环境
	Environment string `json:"environment"`
	// OTLP 配置
	OTLP OTLPConfig `json:"otlp"`
	// Sampler 采样配置
	Sampler SamplerConfig `json:"sampler"`
	// BatchSpan 批量 Span 配置
	BatchSpan BatchSpanConfig `json:"batchSpan"`
}

// OTLPConfig OTLP 导出器配置
type OTLPConfig struct {
	// Endpoint 导出端点，如 localhost:4317
	Endpoint string `json:"endpoint"`
	// Protocol 协议：grpc 或 http
	Protocol string `json:"protocol"`
	// Insecure 是否使用非安全连接
	Insecure bool `json:"insecure"`
	// Headers 自定义请求头
	Headers map[string]string `json:"headers"`
	// Timeout 超时时间
	Timeout time.Duration `json:"timeout"`
	// Compression 压缩方式：gzip 或空
	Compression string `json:"compression"`
}

// SamplerConfig 采样器配置
type SamplerConfig struct {
	// Type 采样类型：always, never, ratio, parentBased
	Type string `json:"type"`
	// Ratio 采样比例（当 Type 为 ratio 时生效）
	Ratio float64 `json:"ratio"`
}

// BatchSpanConfig 批量 Span 处理配置
type BatchSpanConfig struct {
	// MaxQueueSize 队列最大大小
	MaxQueueSize int `json:"maxQueueSize"`
	// MaxExportBatchSize 批量导出最大大小
	MaxExportBatchSize int `json:"maxExportBatchSize"`
	// ExportTimeout 导出超时时间
	ExportTimeout time.Duration `json:"exportTimeout"`
	// ScheduleDelay 调度延迟
	ScheduleDelay time.Duration `json:"scheduleDelay"`
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:        false,
		ServiceName:    "dawn-service",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		OTLP: OTLPConfig{
			Endpoint: "localhost:4317",
			Protocol: "grpc",
			Insecure: true,
			Timeout:  10 * time.Second,
		},
		Sampler: SamplerConfig{
			Type:  "ratio",
			Ratio: 0.1,
		},
		BatchSpan: BatchSpanConfig{
			MaxQueueSize:       2048,
			MaxExportBatchSize: 512,
			ExportTimeout:      30 * time.Second,
			ScheduleDelay:      5 * time.Second,
		},
	}
}

// ============================================================================
// 配置键常量
// ============================================================================

const (
	configKeyEnabled        = "etc.tracing.enabled"
	configKeyServiceName    = "etc.tracing.serviceName"
	configKeyServiceVersion = "etc.tracing.serviceVersion"
	configKeyEnvironment    = "etc.tracing.environment"
	configKeyOTLPEndpoint   = "etc.tracing.otlp.endpoint"
	configKeyOTLPProtocol   = "etc.tracing.otlp.protocol"
	configKeyOTLPInsecure   = "etc.tracing.otlp.insecure"
	configKeyOTLPTimeout    = "etc.tracing.otlp.timeout"
	configKeySamplerType    = "etc.tracing.sampler.type"
	configKeySamplerRatio   = "etc.tracing.sampler.ratio"
)

// LoadConfig 从配置文件加载配置
func LoadConfig() *Config {
	cfg := DefaultConfig()

	cfg.Enabled = etc.Get(configKeyEnabled, cfg.Enabled).Bool()
	cfg.ServiceName = etc.Get(configKeyServiceName, cfg.ServiceName).String()
	cfg.ServiceVersion = etc.Get(configKeyServiceVersion, cfg.ServiceVersion).String()
	cfg.Environment = etc.Get(configKeyEnvironment, cfg.Environment).String()
	cfg.OTLP.Endpoint = etc.Get(configKeyOTLPEndpoint, cfg.OTLP.Endpoint).String()
	cfg.OTLP.Protocol = etc.Get(configKeyOTLPProtocol, cfg.OTLP.Protocol).String()
	cfg.OTLP.Insecure = etc.Get(configKeyOTLPInsecure, cfg.OTLP.Insecure).Bool()
	cfg.OTLP.Timeout = etc.Get(configKeyOTLPTimeout, cfg.OTLP.Timeout).Duration()
	cfg.Sampler.Type = etc.Get(configKeySamplerType, cfg.Sampler.Type).String()
	cfg.Sampler.Ratio = etc.Get(configKeySamplerRatio, cfg.Sampler.Ratio).Float64()

	return cfg
}

// ============================================================================
// 初始化函数
// ============================================================================

// InitTracing 初始化追踪系统
func InitTracing(cfg *Config) (func(context.Context) error, error) {
	if cfg == nil {
		cfg = LoadConfig()
	}

	if !cfg.Enabled {
		log.Info("tracing: disabled")
		return func(ctx context.Context) error { return nil }, nil
	}

	// 创建资源
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// 创建导出器
	exporter, err := createExporter(cfg)
	if err != nil {
		return nil, err
	}

	// 创建采样器
	sampler := createSampler(cfg.Sampler)

	// 创建批量处理器选项
	batchOpts := []sdktrace.BatchSpanProcessorOption{}
	if cfg.BatchSpan.MaxQueueSize > 0 {
		batchOpts = append(batchOpts, sdktrace.WithMaxQueueSize(cfg.BatchSpan.MaxQueueSize))
	}
	if cfg.BatchSpan.MaxExportBatchSize > 0 {
		batchOpts = append(batchOpts, sdktrace.WithMaxExportBatchSize(cfg.BatchSpan.MaxExportBatchSize))
	}
	if cfg.BatchSpan.ExportTimeout > 0 {
		batchOpts = append(batchOpts, sdktrace.WithExportTimeout(cfg.BatchSpan.ExportTimeout))
	}
	if cfg.BatchSpan.ScheduleDelay > 0 {
		batchOpts = append(batchOpts, sdktrace.WithBatchTimeout(cfg.BatchSpan.ScheduleDelay))
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter, batchOpts...),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 注册到 dawn tracing
	tracing.SetTracerProvider(NewProvider(WithTracerProvider(tp)))

	log.Infof("tracing: initialized with OTLP exporter (endpoint=%s, protocol=%s)",
		cfg.OTLP.Endpoint, cfg.OTLP.Protocol)

	// 返回关闭函数
	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

// createExporter 创建 OTLP 导出器
func createExporter(cfg *Config) (*otlptrace.Exporter, error) {
	ctx := context.Background()

	switch cfg.OTLP.Protocol {
	case "http", "HTTP":
		return createHTTPExporter(ctx, cfg)
	default: // grpc
		return createGRPCExporter(ctx, cfg)
	}
}

// createGRPCExporter 创建 gRPC 导出器
func createGRPCExporter(ctx context.Context, cfg *Config) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLP.Endpoint),
	}

	if cfg.OTLP.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	if cfg.OTLP.Timeout > 0 {
		opts = append(opts, otlptracegrpc.WithTimeout(cfg.OTLP.Timeout))
	}

	if len(cfg.OTLP.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.OTLP.Headers))
	}

	if cfg.OTLP.Compression == "gzip" {
		opts = append(opts, otlptracegrpc.WithCompressor("gzip"))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createHTTPExporter 创建 HTTP 导出器
func createHTTPExporter(ctx context.Context, cfg *Config) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.OTLP.Endpoint),
	}

	if cfg.OTLP.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if cfg.OTLP.Timeout > 0 {
		opts = append(opts, otlptracehttp.WithTimeout(cfg.OTLP.Timeout))
	}

	if len(cfg.OTLP.Headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLP.Headers))
	}

	if cfg.OTLP.Compression == "gzip" {
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	}

	return otlptracehttp.New(ctx, opts...)
}

// createSampler 创建采样器
func createSampler(cfg SamplerConfig) sdktrace.Sampler {
	switch cfg.Type {
	case "always":
		return sdktrace.AlwaysSample()
	case "never":
		return sdktrace.NeverSample()
	case "parentBased":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Ratio))
	case "ratio":
		return sdktrace.TraceIDRatioBased(cfg.Ratio)
	default:
		return sdktrace.TraceIDRatioBased(0.1)
	}
}

// ============================================================================
// 便捷函数
// ============================================================================

// MustInitTracing 初始化追踪，失败时 panic
func MustInitTracing(cfg *Config) func(context.Context) error {
	shutdown, err := InitTracing(cfg)
	if err != nil {
		log.Fatalf("tracing: failed to initialize: %v", err)
	}
	return shutdown
}

// InitFromConfig 从配置文件初始化追踪
func InitFromConfig() (func(context.Context) error, error) {
	return InitTracing(LoadConfig())
}

// GetOtelTracerProvider 获取 OpenTelemetry TracerProvider
func GetOtelTracerProvider() oteltrace.TracerProvider {
	return otel.GetTracerProvider()
}
