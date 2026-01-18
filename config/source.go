// Package config 提供统一的配置管理接口和实现。
// 支持从多种配置源（文件、etcd、consul、nacos）加载和监听配置项。
package config

import "context"

const (
	ReadOnly  Mode = "read-only"  // 只读
	WriteOnly Mode = "write-only" // 只写
	ReadWrite Mode = "read-write" // 读写
)

// Mode 配置源的模式：只读、只写或读写
type Mode string

// Source 配置源接口，定义了配置的加载、存储和监听功能。
// 实现该接口可以支持不同的配置后端（文件系统、etcd、consul、nacos 等）。
type Source interface {
	// Name 配置源名称
	Name() string
	// Load 加载配置项
	Load(ctx context.Context, file ...string) ([]*Configuration, error)
	// Store 保存配置项
	Store(ctx context.Context, file string, content []byte) error
	// Watch 监听配置项
	Watch(ctx context.Context) (Watcher, error)
	// Close 关闭配置源
	Close() error
}

// Watcher 配置监听器接口，用于监听配置变化。
type Watcher interface {
	// Next 返回配置列表
	Next() ([]*Configuration, error)
	// Stop 停止监听
	Stop() error
}
