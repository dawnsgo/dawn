package network

import "sync"

// DefaultAttr 默认属性实现
type DefaultAttr struct {
	values sync.Map
}

// NewAttr 创建属性实例
func NewAttr() *DefaultAttr {
	return &DefaultAttr{}
}

// Get 获取属性值
func (a *DefaultAttr) Get(key any) (any, bool) {
	return a.values.Load(key)
}

// Set 设置属性值
func (a *DefaultAttr) Set(key, value any) {
	a.values.Store(key, value)
}

// Del 删除属性值
func (a *DefaultAttr) Del(key any) (ok bool) {
	_, ok = a.values.LoadAndDelete(key)
	return
}

// Visit 访问所有的属性值
func (a *DefaultAttr) Visit(fn func(key, value any) bool) {
	a.values.Range(fn)
}
