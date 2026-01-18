// Package registry 提供服务注册发现的统一接口和实现。
// 支持多种注册中心（etcd、consul、nacos），用于构建微服务架构。
package registry

import (
	"context"
)

// Registry 服务注册中心接口，提供服务的注册、解注册、监听和查询功能。
type Registry interface {
	// Name 获取服务注册发现组件名
	Name() string
	// Register 注册服务实例
	Register(ctx context.Context, ins *ServiceInstance) error
	// Deregister 解注册服务实例
	Deregister(ctx context.Context, ins *ServiceInstance) error
	// Watch 监听相同服务名的服务实例变化
	Watch(ctx context.Context, serviceName string) (Watcher, error)
	// Services 获取服务实例列表
	Services(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
}

// Discovery 服务发现接口，提供服务的监听和查询功能（不包含注册功能）。
type Discovery interface {
	// Watch 监听相同服务名的服务实例变化
	Watch(ctx context.Context, serviceName string) (Watcher, error)
	// Services 获取服务实例列表
	Services(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
}

// Watcher 服务监听器接口，用于监听服务实例的变化。
type Watcher interface {
	// Next 返回服务实例列表
	Next() ([]*ServiceInstance, error)
	// Stop 停止监听
	Stop() error
}

// ServiceInstance 服务实例结构体，表示一个微服务实例的完整信息。
type ServiceInstance struct {
	// 服务实体ID，每个服务实体ID唯一
	ID string `json:"id,omitempty"`
	// 服务实体名
	Name string `json:"name,omitempty"`
	// 服务实体类型
	Kind string `json:"kind,omitempty"`
	// 服务实体别名
	Alias string `json:"alias,omitempty"`
	// 服务实例状态
	State string `json:"state,omitempty"`
	// 服务事件集合
	Events []int `json:"events,omitempty"`
	// 服务路由ID
	Routes []Route `json:"routes,omitempty"`
	// 服务路由列表
	Services []string `json:"services,omitempty"`
	// 微服务实体暴露端口
	Endpoint string `json:"endpoint,omitempty"`
	// 微服务路由加权轮询权重
	Weight int `json:"weight,omitempty"`
	// 元数据
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Route 路由信息结构体，定义了服务路由的配置。
type Route struct {
	// 路由ID
	ID int32 `json:"i,omitempty"`
	// 是否内部路由
	Internal bool `json:"n,omitempty"`
	// 是否有状态路由
	Stateful bool `json:"s,omitempty"`
	// 是否授权路由
	Authorized bool `json:"a,omitempty"`
}
