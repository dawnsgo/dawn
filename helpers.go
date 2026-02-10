package dawn

// GetContext 获取默认上下文（返回接口类型，便于 mock）
func GetContext() Contextor {
	return Default()
}
