/**
 * @Author: dawn
 * @Desc: Builtin Metrics 单元测试
 */

package metrics

import (
	"testing"
)

func TestInitBuiltinMetrics(t *testing.T) {
	// 初始化内置指标
	InitBuiltinMetrics()

	// 验证指标已创建
	if NetworkConnectionsTotal == nil {
		t.Fatal("NetworkConnectionsTotal should be initialized")
	}
	if NetworkConnectionsActive == nil {
		t.Fatal("NetworkConnectionsActive should be initialized")
	}
	if NetworkBytesReceived == nil {
		t.Fatal("NetworkBytesReceived should be initialized")
	}
	if NetworkBytesSent == nil {
		t.Fatal("NetworkBytesSent should be initialized")
	}
	if NetworkMessagesReceived == nil {
		t.Fatal("NetworkMessagesReceived should be initialized")
	}
	if NetworkMessagesSent == nil {
		t.Fatal("NetworkMessagesSent should be initialized")
	}
	if NetworkMessageDuration == nil {
		t.Fatal("NetworkMessageDuration should be initialized")
	}

	if SessionTotal == nil {
		t.Fatal("SessionTotal should be initialized")
	}
	if SessionActive == nil {
		t.Fatal("SessionActive should be initialized")
	}
	if SessionBindTotal == nil {
		t.Fatal("SessionBindTotal should be initialized")
	}

	if GateRequestsTotal == nil {
		t.Fatal("GateRequestsTotal should be initialized")
	}
	if GateRequestDuration == nil {
		t.Fatal("GateRequestDuration should be initialized")
	}
	if GateUpstreamDuration == nil {
		t.Fatal("GateUpstreamDuration should be initialized")
	}

	if NodeRequestsTotal == nil {
		t.Fatal("NodeRequestsTotal should be initialized")
	}
	if NodeRequestDuration == nil {
		t.Fatal("NodeRequestDuration should be initialized")
	}
	if NodeActorsActive == nil {
		t.Fatal("NodeActorsActive should be initialized")
	}

	if HTTPRequestsTotal == nil {
		t.Fatal("HTTPRequestsTotal should be initialized")
	}
	if HTTPRequestDuration == nil {
		t.Fatal("HTTPRequestDuration should be initialized")
	}
	if HTTPRequestSize == nil {
		t.Fatal("HTTPRequestSize should be initialized")
	}
	if HTTPResponseSize == nil {
		t.Fatal("HTTPResponseSize should be initialized")
	}
}

func TestBuiltinMetrics_Usage(t *testing.T) {
	InitBuiltinMetrics()

	// 测试网络指标
	NetworkConnectionsTotal.WithLabelValues("instance1", "tcp").Inc()
	NetworkConnectionsActive.WithLabelValues("instance1", "tcp").Inc()
	NetworkBytesReceived.WithLabelValues("instance1", "tcp").Add(1024)
	NetworkBytesSent.WithLabelValues("instance1", "tcp").Add(2048)
	NetworkMessagesReceived.WithLabelValues("instance1", "tcp", "route1").Inc()
	NetworkMessagesSent.WithLabelValues("instance1", "tcp", "route1").Inc()
	NetworkMessageDuration.WithLabelValues("instance1", "tcp", "route1").Observe(10.5)

	// 测试会话指标
	SessionTotal.WithLabelValues("instance1", "conn").Inc()
	SessionActive.WithLabelValues("instance1", "conn").Set(10)
	SessionBindTotal.WithLabelValues("instance1", "ok").Inc()

	// 测试 Gate 指标
	GateRequestsTotal.WithLabelValues("instance1", "route1", "200").Inc()
	GateRequestDuration.WithLabelValues("instance1", "route1").Observe(15.5)
	GateUpstreamDuration.WithLabelValues("instance1", "route1").Observe(20.0)

	// 测试 Node 指标
	NodeRequestsTotal.WithLabelValues("instance1", "route1", "200").Inc()
	NodeRequestDuration.WithLabelValues("instance1", "route1").Observe(12.5)
	NodeActorsActive.WithLabelValues("instance1").Set(5)

	// 测试 HTTP 指标
	HTTPRequestsTotal.WithLabelValues("instance1", "GET", "/api", "200").Inc()
	HTTPRequestDuration.WithLabelValues("instance1", "GET", "/api").Observe(8.5)
	HTTPRequestSize.WithLabelValues("instance1", "GET", "/api").Observe(256)
	HTTPResponseSize.WithLabelValues("instance1", "GET", "/api").Observe(512)
}
