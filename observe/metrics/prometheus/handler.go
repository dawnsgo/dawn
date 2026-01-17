/**
 * @Author: dawn
 * @Desc: Prometheus HTTP 处理器
 */

package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler 返回 Prometheus HTTP 处理器
func Handler() http.Handler {
	return promhttp.Handler()
}

// HandlerFor 返回指定 Registry 的 Prometheus HTTP 处理器
func HandlerFor(registry *prometheus.Registry) http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
