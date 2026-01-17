/**
 * @Author: dawn
 * @Desc: 框架内置指标定义
 */

package metrics

// ============================================================================
// 命名空间和子系统
// ============================================================================

const (
	// Namespace 框架命名空间
	Namespace = "dawn"

	// 子系统
	SubsystemNetwork = "network"
	SubsystemGate    = "gate"
	SubsystemNode    = "node"
	SubsystemMesh    = "mesh"
	SubsystemSession = "session"
	SubsystemHTTP    = "http"
)

// ============================================================================
// 标签名
// ============================================================================

const (
	LabelInstance  = "instance"
	LabelProtocol  = "protocol"
	LabelMethod    = "method"
	LabelRoute     = "route"
	LabelStatus    = "status"
	LabelCode      = "code"
	LabelKind      = "kind"
	LabelDirection = "direction"
)

// ============================================================================
// 网络层指标
// ============================================================================

var (
	// NetworkConnectionsTotal 网络连接总数
	NetworkConnectionsTotal CounterVec

	// NetworkConnectionsActive 当前活跃连接数
	NetworkConnectionsActive GaugeVec

	// NetworkBytesReceived 接收字节数
	NetworkBytesReceived CounterVec

	// NetworkBytesSent 发送字节数
	NetworkBytesSent CounterVec

	// NetworkMessagesReceived 接收消息数
	NetworkMessagesReceived CounterVec

	// NetworkMessagesSent 发送消息数
	NetworkMessagesSent CounterVec

	// NetworkMessageDuration 消息处理延迟（毫秒）
	NetworkMessageDuration HistogramVec
)

// ============================================================================
// 会话层指标
// ============================================================================

var (
	// SessionTotal 会话总数
	SessionTotal CounterVec

	// SessionActive 当前活跃会话数
	SessionActive GaugeVec

	// SessionBindTotal 绑定操作总数
	SessionBindTotal CounterVec
)

// ============================================================================
// Gate 指标
// ============================================================================

var (
	// GateRequestsTotal 请求总数
	GateRequestsTotal CounterVec

	// GateRequestDuration 请求延迟（毫秒）
	GateRequestDuration HistogramVec

	// GateUpstreamDuration 上游延迟（毫秒）
	GateUpstreamDuration HistogramVec
)

// ============================================================================
// Node 指标
// ============================================================================

var (
	// NodeRequestsTotal 请求总数
	NodeRequestsTotal CounterVec

	// NodeRequestDuration 请求延迟（毫秒）
	NodeRequestDuration HistogramVec

	// NodeActorsActive 当前活跃 Actor 数
	NodeActorsActive GaugeVec
)

// ============================================================================
// HTTP 指标
// ============================================================================

var (
	// HTTPRequestsTotal HTTP 请求总数
	HTTPRequestsTotal CounterVec

	// HTTPRequestDuration HTTP 请求延迟（毫秒）
	HTTPRequestDuration HistogramVec

	// HTTPRequestSize HTTP 请求大小（字节）
	HTTPRequestSize HistogramVec

	// HTTPResponseSize HTTP 响应大小（字节）
	HTTPResponseSize HistogramVec
)

// ============================================================================
// 初始化函数
// ============================================================================

// InitBuiltinMetrics 初始化内置指标
// 应在设置 Provider 之后调用
func InitBuiltinMetrics() {
	initNetworkMetrics()
	initSessionMetrics()
	initGateMetrics()
	initNodeMetrics()
	initHTTPMetrics()
}

func initNetworkMetrics() {
	NetworkConnectionsTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "connections_total",
		Help:      "Total number of network connections",
	}, []string{LabelInstance, LabelProtocol})

	NetworkConnectionsActive = NewGaugeVec(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "connections_active",
		Help:      "Number of active network connections",
	}, []string{LabelInstance, LabelProtocol})

	NetworkBytesReceived = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "bytes_received_total",
		Help:      "Total bytes received",
	}, []string{LabelInstance, LabelProtocol})

	NetworkBytesSent = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "bytes_sent_total",
		Help:      "Total bytes sent",
	}, []string{LabelInstance, LabelProtocol})

	NetworkMessagesReceived = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "messages_received_total",
		Help:      "Total messages received",
	}, []string{LabelInstance, LabelProtocol, LabelRoute})

	NetworkMessagesSent = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "messages_sent_total",
		Help:      "Total messages sent",
	}, []string{LabelInstance, LabelProtocol, LabelRoute})

	NetworkMessageDuration = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNetwork,
		Name:      "message_duration_milliseconds",
		Help:      "Message processing duration in milliseconds",
		Buckets:   DefaultBuckets,
	}, []string{LabelInstance, LabelProtocol, LabelRoute})
}

func initSessionMetrics() {
	SessionTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemSession,
		Name:      "total",
		Help:      "Total number of sessions created",
	}, []string{LabelInstance, LabelKind})

	SessionActive = NewGaugeVec(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemSession,
		Name:      "active",
		Help:      "Number of active sessions",
	}, []string{LabelInstance, LabelKind})

	SessionBindTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemSession,
		Name:      "bind_total",
		Help:      "Total number of session bind operations",
	}, []string{LabelInstance, LabelStatus})
}

func initGateMetrics() {
	GateRequestsTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemGate,
		Name:      "requests_total",
		Help:      "Total number of gate requests",
	}, []string{LabelInstance, LabelRoute, LabelCode})

	GateRequestDuration = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemGate,
		Name:      "request_duration_milliseconds",
		Help:      "Gate request duration in milliseconds",
		Buckets:   DefaultBuckets,
	}, []string{LabelInstance, LabelRoute})

	GateUpstreamDuration = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemGate,
		Name:      "upstream_duration_milliseconds",
		Help:      "Gate upstream duration in milliseconds",
		Buckets:   DefaultBuckets,
	}, []string{LabelInstance, LabelRoute})
}

func initNodeMetrics() {
	NodeRequestsTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNode,
		Name:      "requests_total",
		Help:      "Total number of node requests",
	}, []string{LabelInstance, LabelRoute, LabelCode})

	NodeRequestDuration = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNode,
		Name:      "request_duration_milliseconds",
		Help:      "Node request duration in milliseconds",
		Buckets:   DefaultBuckets,
	}, []string{LabelInstance, LabelRoute})

	NodeActorsActive = NewGaugeVec(GaugeOpts{
		Namespace: Namespace,
		Subsystem: SubsystemNode,
		Name:      "actors_active",
		Help:      "Number of active actors",
	}, []string{LabelInstance})
}

func initHTTPMetrics() {
	HTTPRequestsTotal = NewCounterVec(CounterOpts{
		Namespace: Namespace,
		Subsystem: SubsystemHTTP,
		Name:      "requests_total",
		Help:      "Total number of HTTP requests",
	}, []string{LabelInstance, LabelMethod, LabelRoute, LabelStatus})

	HTTPRequestDuration = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemHTTP,
		Name:      "request_duration_milliseconds",
		Help:      "HTTP request duration in milliseconds",
		Buckets:   DefaultBuckets,
	}, []string{LabelInstance, LabelMethod, LabelRoute})

	HTTPRequestSize = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemHTTP,
		Name:      "request_size_bytes",
		Help:      "HTTP request size in bytes",
		Buckets:   DefaultByteBuckets,
	}, []string{LabelInstance, LabelMethod, LabelRoute})

	HTTPResponseSize = NewHistogramVec(HistogramOpts{
		Namespace: Namespace,
		Subsystem: SubsystemHTTP,
		Name:      "response_size_bytes",
		Help:      "HTTP response size in bytes",
		Buckets:   DefaultByteBuckets,
	}, []string{LabelInstance, LabelMethod, LabelRoute})
}
