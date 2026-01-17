/**
 * @Author: dawn
 * @Desc: Context 单元测试
 */

package dawn

import (
	"testing"
)

// mockLogger 模拟日志记录器
type mockLogger struct{}

func (m *mockLogger) Close() error { return nil }

// mockConfigurator 模拟配置器
type mockConfigurator struct{}

func (m *mockConfigurator) Close() error { return nil }

// mockEventbus 模拟事件总线
type mockEventbus struct{}

func (m *mockEventbus) Close() error { return nil }

// mockTaskPool 模拟任务池
type mockTaskPool struct{}

func (m *mockTaskPool) Release() {}

// mockCache 模拟缓存
type mockCache struct{}

func (m *mockCache) Close() error { return nil }

// mockLockMaker 模拟锁制造器
type mockLockMaker struct{}

func (m *mockLockMaker) Close() error { return nil }

func TestNewContext(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}
}

func TestDefault(t *testing.T) {
	ctx1 := Default()
	ctx2 := Default()

	if ctx1 != ctx2 {
		t.Fatal("Default() should return the same instance")
	}
}

func TestContext_SetLogger(t *testing.T) {
	ctx := NewContext()
	logger := &mockLogger{}

	ctx.SetLogger(logger)
	if ctx.Logger() != logger {
		t.Fatal("Logger not set correctly")
	}

	// 测试设置 nil
	ctx.SetLogger(nil)
	if ctx.Logger() != logger {
		t.Fatal("Logger should not change when setting nil")
	}

	// 测试替换
	logger2 := &mockLogger{}
	ctx.SetLogger(logger2)
	if ctx.Logger() != logger2 {
		t.Fatal("Logger not replaced correctly")
	}
}

func TestContext_SetConfigurator(t *testing.T) {
	ctx := NewContext()
	configurator := &mockConfigurator{}

	ctx.SetConfigurator(configurator)
	if ctx.Configurator() != configurator {
		t.Fatal("Configurator not set correctly")
	}

	// 测试设置 nil
	ctx.SetConfigurator(nil)
	if ctx.Configurator() != configurator {
		t.Fatal("Configurator should not change when setting nil")
	}

	// 测试替换
	configurator2 := &mockConfigurator{}
	ctx.SetConfigurator(configurator2)
	if ctx.Configurator() != configurator2 {
		t.Fatal("Configurator not replaced correctly")
	}
}

func TestContext_SetEventbus(t *testing.T) {
	ctx := NewContext()
	eventbus := &mockEventbus{}

	ctx.SetEventbus(eventbus)
	if ctx.Eventbus() != eventbus {
		t.Fatal("Eventbus not set correctly")
	}

	// 测试设置 nil
	ctx.SetEventbus(nil)
	if ctx.Eventbus() != eventbus {
		t.Fatal("Eventbus should not change when setting nil")
	}

	// 测试替换
	eventbus2 := &mockEventbus{}
	ctx.SetEventbus(eventbus2)
	if ctx.Eventbus() != eventbus2 {
		t.Fatal("Eventbus not replaced correctly")
	}
}

func TestContext_SetTaskPool(t *testing.T) {
	ctx := NewContext()
	taskPool := &mockTaskPool{}

	ctx.SetTaskPool(taskPool)
	if ctx.TaskPool() != taskPool {
		t.Fatal("TaskPool not set correctly")
	}

	// 测试设置 nil
	ctx.SetTaskPool(nil)
	if ctx.TaskPool() != taskPool {
		t.Fatal("TaskPool should not change when setting nil")
	}

	// 测试替换
	taskPool2 := &mockTaskPool{}
	ctx.SetTaskPool(taskPool2)
	if ctx.TaskPool() != taskPool2 {
		t.Fatal("TaskPool not replaced correctly")
	}
}

func TestContext_SetCache(t *testing.T) {
	ctx := NewContext()
	cache := &mockCache{}

	ctx.SetCache(cache)
	if ctx.Cache() != cache {
		t.Fatal("Cache not set correctly")
	}

	// 测试设置 nil
	ctx.SetCache(nil)
	if ctx.Cache() != cache {
		t.Fatal("Cache should not change when setting nil")
	}

	// 测试替换
	cache2 := &mockCache{}
	ctx.SetCache(cache2)
	if ctx.Cache() != cache2 {
		t.Fatal("Cache not replaced correctly")
	}
}

func TestContext_SetLockMaker(t *testing.T) {
	ctx := NewContext()
	lockMaker := &mockLockMaker{}

	ctx.SetLockMaker(lockMaker)
	if ctx.LockMaker() != lockMaker {
		t.Fatal("LockMaker not set correctly")
	}

	// 测试设置 nil
	ctx.SetLockMaker(nil)
	if ctx.LockMaker() != lockMaker {
		t.Fatal("LockMaker should not change when setting nil")
	}

	// 测试替换
	lockMaker2 := &mockLockMaker{}
	ctx.SetLockMaker(lockMaker2)
	if ctx.LockMaker() != lockMaker2 {
		t.Fatal("LockMaker not replaced correctly")
	}
}

func TestContext_Close(t *testing.T) {
	ctx := NewContext()

	// 设置所有资源
	ctx.SetLogger(&mockLogger{})
	ctx.SetConfigurator(&mockConfigurator{})
	ctx.SetEventbus(&mockEventbus{})
	ctx.SetTaskPool(&mockTaskPool{})
	ctx.SetCache(&mockCache{})
	ctx.SetLockMaker(&mockLockMaker{})

	// 关闭上下文
	err := ctx.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// 验证所有资源都被清空
	if ctx.Logger() != nil {
		t.Fatal("Logger should be nil after Close()")
	}
	if ctx.Configurator() != nil {
		t.Fatal("Configurator should be nil after Close()")
	}
	if ctx.Eventbus() != nil {
		t.Fatal("Eventbus should be nil after Close()")
	}
	if ctx.TaskPool() != nil {
		t.Fatal("TaskPool should be nil after Close()")
	}
	if ctx.Cache() != nil {
		t.Fatal("Cache should be nil after Close()")
	}
	if ctx.LockMaker() != nil {
		t.Fatal("LockMaker should be nil after Close()")
	}
}
