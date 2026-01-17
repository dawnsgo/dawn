/**
 * @Author: dawn
 * @Desc: Metrics 组件选项
 */

package metrics

type options struct {
	addr                 string
	path                 string
	enableGoMetrics      bool
	enableProcessMetrics bool
}

func defaultOptions() *options {
	return &options{
		addr:                 ":9090",
		path:                 "/metrics",
		enableGoMetrics:      true,
		enableProcessMetrics: true,
	}
}

// Option 配置选项函数
type Option func(*options)

// WithAddr 设置监听地址
func WithAddr(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

// WithPath 设置路径
func WithPath(path string) Option {
	return func(o *options) {
		o.path = path
	}
}

// WithGoMetrics 启用 Go 运行时指标
func WithGoMetrics(enable bool) Option {
	return func(o *options) {
		o.enableGoMetrics = enable
	}
}

// WithProcessMetrics 启用进程指标
func WithProcessMetrics(enable bool) Option {
	return func(o *options) {
		o.enableProcessMetrics = enable
	}
}
