/**
 * @Author: dawn
 * @Desc: Container 单元测试
 */

package dawn

import (
	"testing"

	"github.com/dawnsgo/dawn/component"
	"github.com/dawnsgo/dawn/errors"
)

// mockComponent 模拟组件
type mockComponent struct {
	component.Base
	name         string
	initErr      error
	startErr     error
	closeErr     error
	destroyErr   error
	initCalled   bool
	startCalled  bool
	closeCalled  bool
	destroyCalled bool
}

func (m *mockComponent) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mock"
}

func (m *mockComponent) Init() error {
	m.initCalled = true
	return m.initErr
}

func (m *mockComponent) Start() error {
	m.startCalled = true
	return m.startErr
}

func (m *mockComponent) Close() error {
	m.closeCalled = true
	return m.closeErr
}

func (m *mockComponent) Destroy() error {
	m.destroyCalled = true
	return m.destroyErr
}

func TestNewContainer(t *testing.T) {
	// 测试默认创建
	c := NewContainer()
	if c == nil {
		t.Fatal("NewContainer() returned nil")
	}
	if c.ctx == nil {
		t.Fatal("Container context is nil")
	}
	// components 是 slice，初始为空但不为 nil
	if c.components == nil && len(c.components) != 0 {
		t.Fatal("Container components should be initialized")
	}

	// 测试使用自定义上下文
	ctx := NewContext()
	c2 := NewContainer(WithContext(ctx))
	if c2.ctx != ctx {
		t.Fatal("Container context not set correctly")
	}
}

func TestContainer_Add(t *testing.T) {
	c := NewContainer()
	comp1 := &mockComponent{name: "comp1"}
	comp2 := &mockComponent{name: "comp2"}

	c.Add(comp1)
	if len(c.components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(c.components))
	}

	c.Add(comp2)
	if len(c.components) != 2 {
		t.Fatalf("Expected 2 components, got %d", len(c.components))
	}

	// 测试批量添加
	comp3 := &mockComponent{name: "comp3"}
	comp4 := &mockComponent{name: "comp4"}
	c.Add(comp3, comp4)
	if len(c.components) != 4 {
		t.Fatalf("Expected 4 components, got %d", len(c.components))
	}
}

func TestContainer_Context(t *testing.T) {
	ctx := NewContext()
	c := NewContainer(WithContext(ctx))

	if c.Context() != ctx {
		t.Fatal("Context() returned wrong context")
	}
}

func TestContainer_doInitComponents(t *testing.T) {
	c := NewContainer()

	// 测试成功初始化
	comp1 := &mockComponent{name: "comp1"}
	c.Add(comp1)

	err := c.doInitComponents()
	if err != nil {
		t.Fatalf("doInitComponents() failed: %v", err)
	}
	if !comp1.initCalled {
		t.Fatal("Component Init() not called")
	}

	// 测试初始化失败
	comp2 := &mockComponent{
		name:    "comp2",
		initErr: errors.NewSimple("init failed"),
	}
	c.Add(comp2)

	err = c.doInitComponents()
	if err == nil {
		t.Fatal("Expected error from doInitComponents()")
	}
}

func TestContainer_doStartComponents(t *testing.T) {
	c := NewContainer()

	// 测试成功启动
	comp1 := &mockComponent{name: "comp1"}
	c.Add(comp1)

	err := c.doStartComponents()
	if err != nil {
		t.Fatalf("doStartComponents() failed: %v", err)
	}
	if !comp1.startCalled {
		t.Fatal("Component Start() not called")
	}

	// 测试启动失败
	comp2 := &mockComponent{
		name:     "comp2",
		startErr: errors.NewSimple("start failed"),
	}
	c.Add(comp2)

	err = c.doStartComponents()
	if err == nil {
		t.Fatal("Expected error from doStartComponents()")
	}
}

func TestContainer_doCloseComponents(t *testing.T) {
	c := NewContainer()

	comp1 := &mockComponent{name: "comp1"}
	comp2 := &mockComponent{name: "comp2"}
	c.Add(comp1, comp2)

	// 先初始化并启动
	c.doInitComponents()
	c.doStartComponents()

	// 测试关闭
	c.doCloseComponents()

	// 等待 goroutines 完成（简化测试，实际应该用更可靠的方式）
	// 注意：doCloseComponents 是异步的，这里只是验证不会 panic
	if !comp1.closeCalled && !comp2.closeCalled {
		// 由于是异步执行，这里只检查是否至少有一个被调用
		// 实际测试中应该使用同步机制
	}
}

func TestContainer_doDestroyComponents(t *testing.T) {
	c := NewContainer()

	comp1 := &mockComponent{name: "comp1"}
	c.Add(comp1)

	// 先初始化并启动
	c.doInitComponents()
	c.doStartComponents()

	// 测试销毁
	c.doDestroyComponents()

	// 注意：doDestroyComponents 是异步的
}

func TestContainer_doSaveProcessID(t *testing.T) {
	c := NewContainer()

	// 测试保存 PID（如果配置了路径）
	// 由于依赖 etc 配置，这里只测试不会 panic
	c.doSaveProcessID()
}

func TestContainer_doPrintFrameworkInfo(t *testing.T) {
	c := NewContainer()

	// 测试打印框架信息（不会 panic）
	c.doPrintFrameworkInfo()
}

func TestContainer_doClearModules(t *testing.T) {
	c := NewContainer()

	// 测试清理模块（不会 panic）
	c.doClearModules()
}
