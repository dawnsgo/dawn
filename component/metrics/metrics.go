/**
 * @Author: dawn
 * @Desc: Metrics HTTP 服务组件
 */

package metrics

import (
	"fmt"
	"net/http"

	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/component"
	"github.com/dawnsgo/dawn/core/info"
	xnet "github.com/dawnsgo/dawn/core/net"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/observe/metrics"
	promadapter "github.com/dawnsgo/dawn/observe/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var _ component.Component = &Server{}

// Server Metrics HTTP 服务组件
type Server struct {
	component.Base
	opts     *options
	registry *prometheus.Registry
}

// NewServer 创建 Metrics 服务器
func NewServer(opts ...Option) *Server {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &Server{
		opts:     o,
		registry: prometheus.NewRegistry(),
	}
}

// Name 组件名称
func (s *Server) Name() string {
	return "metrics"
}

// Init 初始化
func (s *Server) Init() error {
	// 注册默认收集器
	if s.opts.enableGoMetrics {
		s.registry.MustRegister(prometheus.NewGoCollector())
	}
	if s.opts.enableProcessMetrics {
		s.registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}

	// 设置全局 Provider
	provider := promadapter.NewProvider(promadapter.WithRegistry(s.registry))
	metrics.SetProvider(provider)

	// 初始化内置指标
	metrics.InitBuiltinMetrics()

	return nil
}

// Start 启动服务
func (s *Server) Start() error {
	listenAddr, exposeAddr, err := xnet.ParseAddr(s.opts.addr)
	if err != nil {
		return errors.WrapWithCode(err, codes.InvalidConfig, "metrics addr parse failed")
	}

	mux := http.NewServeMux()
	mux.Handle(s.opts.path, promadapter.HandlerFor(s.registry))

	go func() {
		if err := http.ListenAndServe(listenAddr, mux); err != nil {
			log.Errorf("metrics server start failed: %v", err)
		}
	}()

	info.PrintBoxInfo("Metrics",
		fmt.Sprintf("Url: http://%s%s", exposeAddr, s.opts.path),
	)

	return nil
}

// Registry 获取 Prometheus Registry
func (s *Server) Registry() *prometheus.Registry {
	return s.registry
}
