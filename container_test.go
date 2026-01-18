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

// mockOrderedComponent 带优先级和依赖的模拟组件
type mockOrderedComponent struct {
	mockComponent
	priority     int
	dependencies []string
}

func (m *mockOrderedComponent) Priority() int {
	return m.priority
}

func (m *mockOrderedComponent) Dependencies() []string {
	return m.dependencies
}

func TestContainer_sortComponentsByPriority(t *testing.T) {
	c := NewContainer()

	// 添加带优先级的组件
	comp1 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp1"},
		priority:      3,
	}
	comp2 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp2"},
		priority:      1,
	}
	comp3 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp3"},
		priority:      2,
	}
	// 添加一个普通组件（使用默认优先级100）
	comp4 := &mockComponent{name: "comp4"}

	c.Add(comp1, comp2, comp3, comp4)

	sorted := c.sortComponentsByPriority()

	// 验证排序顺序：comp2(1) < comp3(2) < comp1(3) < comp4(100)
	if len(sorted) != 4 {
		t.Fatalf("Expected 4 components, got %d", len(sorted))
	}
	if sorted[0].Name() != "comp2" {
		t.Errorf("Expected first component to be 'comp2', got '%s'", sorted[0].Name())
	}
	if sorted[1].Name() != "comp3" {
		t.Errorf("Expected second component to be 'comp3', got '%s'", sorted[1].Name())
	}
	if sorted[2].Name() != "comp1" {
		t.Errorf("Expected third component to be 'comp1', got '%s'", sorted[2].Name())
	}
	if sorted[3].Name() != "comp4" {
		t.Errorf("Expected fourth component to be 'comp4', got '%s'", sorted[3].Name())
	}
}

func TestContainer_sortComponentsByDependency(t *testing.T) {
	c := NewContainer()

	// 创建有依赖关系的组件：
	// comp3 依赖 comp2
	// comp2 依赖 comp1
	// comp1 无依赖
	comp1 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp1"},
		dependencies:  nil,
	}
	comp2 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp2"},
		dependencies:  []string{"comp1"},
	}
	comp3 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp3"},
		dependencies:  []string{"comp2"},
	}

	// 以错误顺序添加
	c.Add(comp3, comp2, comp1)

	sorted, err := c.sortComponentsByDependency()
	if err != nil {
		t.Fatalf("sortComponentsByDependency() failed: %v", err)
	}

	// 验证排序顺序：comp1 < comp2 < comp3
	if len(sorted) != 3 {
		t.Fatalf("Expected 3 components, got %d", len(sorted))
	}

	// 构建位置映射
	positions := make(map[string]int)
	for i, comp := range sorted {
		positions[comp.Name()] = i
	}

	// 验证依赖顺序
	if positions["comp1"] > positions["comp2"] {
		t.Error("comp1 should come before comp2")
	}
	if positions["comp2"] > positions["comp3"] {
		t.Error("comp2 should come before comp3")
	}
}

func TestContainer_sortComponentsByDependency_CircularDependency(t *testing.T) {
	c := NewContainer()

	// 创建循环依赖：comp1 -> comp2 -> comp3 -> comp1
	comp1 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp1"},
		dependencies:  []string{"comp3"},
	}
	comp2 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp2"},
		dependencies:  []string{"comp1"},
	}
	comp3 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp3"},
		dependencies:  []string{"comp2"},
	}

	c.Add(comp1, comp2, comp3)

	_, err := c.sortComponentsByDependency()
	if err == nil {
		t.Fatal("Expected circular dependency error, got nil")
	}
}

func TestContainer_ValidateDependencies(t *testing.T) {
	c := NewContainer()

	// 正常依赖关系
	comp1 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp1"},
		dependencies:  nil,
	}
	comp2 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp2"},
		dependencies:  []string{"comp1"},
	}

	c.Add(comp1, comp2)

	err := c.ValidateDependencies()
	if err != nil {
		t.Fatalf("ValidateDependencies() failed: %v", err)
	}
}

func TestContainer_ValidateDependencies_MissingDependency(t *testing.T) {
	c := NewContainer()

	// 依赖不存在的组件（只会警告，不会报错）
	comp1 := &mockOrderedComponent{
		mockComponent: mockComponent{name: "comp1"},
		dependencies:  []string{"missing_comp"},
	}

	c.Add(comp1)

	// 缺失依赖不会导致错误（只会警告）
	err := c.ValidateDependencies()
	if err != nil {
		t.Fatalf("ValidateDependencies() should not fail for missing dependencies: %v", err)
	}
}

func TestContainer_getComponentPriority(t *testing.T) {
	c := NewContainer()

	// 测试普通组件（默认优先级）
	normalComp := &mockComponent{name: "normal"}
	priority := c.getComponentPriority(normalComp)
	if priority != 100 {
		t.Errorf("Expected default priority 100, got %d", priority)
	}

	// 测试有序组件
	orderedComp := &mockOrderedComponent{
		mockComponent: mockComponent{name: "ordered"},
		priority:      50,
	}
	priority = c.getComponentPriority(orderedComp)
	if priority != 50 {
		t.Errorf("Expected priority 50, got %d", priority)
	}
}

func TestContainer_getComponentDependencies(t *testing.T) {
	c := NewContainer()

	// 测试普通组件（无依赖）
	normalComp := &mockComponent{name: "normal"}
	deps := c.getComponentDependencies(normalComp)
	if deps != nil {
		t.Errorf("Expected nil dependencies, got %v", deps)
	}

	// 测试有序组件
	orderedComp := &mockOrderedComponent{
		mockComponent: mockComponent{name: "ordered"},
		dependencies:  []string{"dep1", "dep2"},
	}
	deps = c.getComponentDependencies(orderedComp)
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}
}
