/**
 * @Author: dawn
 * @Desc: Component 单元测试
 */

package component

import (
	"testing"
)

func TestBase_Name(t *testing.T) {
	b := &Base{}
	if b.Name() != "base" {
		t.Fatalf("Expected 'base', got '%s'", b.Name())
	}
}

func TestBase_Init(t *testing.T) {
	b := &Base{}
	if err := b.Init(); err != nil {
		t.Fatalf("Base.Init() should return nil, got %v", err)
	}
}

func TestBase_Start(t *testing.T) {
	b := &Base{}
	if err := b.Start(); err != nil {
		t.Fatalf("Base.Start() should return nil, got %v", err)
	}
}

func TestBase_Close(t *testing.T) {
	b := &Base{}
	if err := b.Close(); err != nil {
		t.Fatalf("Base.Close() should return nil, got %v", err)
	}
}

func TestBase_Destroy(t *testing.T) {
	b := &Base{}
	if err := b.Destroy(); err != nil {
		t.Fatalf("Base.Destroy() should return nil, got %v", err)
	}
}

// 测试组件接口实现
type testComponent struct {
	Base
	name string
}

func (t *testComponent) Name() string {
	return t.name
}

func TestComponent_Interface(t *testing.T) {
	comp := &testComponent{name: "test"}

	if comp.Name() != "test" {
		t.Fatalf("Expected 'test', got '%s'", comp.Name())
	}

	// 测试基类方法
	if err := comp.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	if err := comp.Start(); err != nil {
		t.Fatalf("Start() failed: %v", err)
	}
	if err := comp.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}
	if err := comp.Destroy(); err != nil {
		t.Fatalf("Destroy() failed: %v", err)
	}
}
