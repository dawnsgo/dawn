package component

// Component 组件接口
type Component interface {
	// Name 组件名称
	Name() string
	// Init 初始化组件
	Init() error
	// Start 启动组件
	Start() error
	// Close 关闭组件
	Close() error
	// Destroy 销毁组件
	Destroy() error
}

// Base 组件基类，提供默认实现
// 嵌入此基类的组件只需实现需要的方法
type Base struct {
}

// Name 组件名称
func (b *Base) Name() string { return "base" }

// Init 初始化组件
func (b *Base) Init() error { return nil }

// Start 启动组件
func (b *Base) Start() error { return nil }

// Close 关闭组件
func (b *Base) Close() error { return nil }

// Destroy 销毁组件
func (b *Base) Destroy() error { return nil }
