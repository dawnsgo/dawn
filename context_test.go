/**
 * @Author: dawn
 * @Desc: Context 单元测试
 */

package dawn

import (
	"context"
	"testing"
	"time"

	"github.com/dawnsgo/dawn/cache"
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/core/value"
	"github.com/dawnsgo/dawn/eventbus"
	dawnlog "github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/lock"
	"github.com/dawnsgo/dawn/task"
)

// ==================== Mock Logger ====================

type mockLogger struct{}

func (m *mockLogger) Print(level dawnlog.Level, a ...any)              {}
func (m *mockLogger) Printf(level dawnlog.Level, format string, a ...any) {}
func (m *mockLogger) Debug(a ...any)                                   {}
func (m *mockLogger) Debugf(format string, a ...any)                   {}
func (m *mockLogger) Info(a ...any)                                    {}
func (m *mockLogger) Infof(format string, a ...any)                    {}
func (m *mockLogger) Warn(a ...any)                                    {}
func (m *mockLogger) Warnf(format string, a ...any)                    {}
func (m *mockLogger) Error(a ...any)                                   {}
func (m *mockLogger) Errorf(format string, a ...any)                   {}
func (m *mockLogger) Fatal(a ...any)                                   {}
func (m *mockLogger) Fatalf(format string, a ...any)                   {}
func (m *mockLogger) Panic(a ...any)                                   {}
func (m *mockLogger) Panicf(format string, a ...any)                   {}
func (m *mockLogger) Close() error                                     { return nil }

// ==================== Mock Configurator ====================

type mockConfigurator struct{}

func (m *mockConfigurator) Has(pattern string) bool                     { return false }
func (m *mockConfigurator) Get(pattern string, def ...any) value.Value  { return value.NewValue() }
func (m *mockConfigurator) Set(pattern string, value any) error         { return nil }
func (m *mockConfigurator) Match(patterns ...string) config.Matcher     { return nil }
func (m *mockConfigurator) Watch(cb config.WatchCallbackFunc, names ...string) {}
func (m *mockConfigurator) Load(ctx context.Context, source string, file ...string) ([]*config.Configuration, error) {
	return nil, nil
}
func (m *mockConfigurator) Store(ctx context.Context, source string, file string, content any, override ...bool) error {
	return nil
}
func (m *mockConfigurator) Close() {}

// ==================== Mock Eventbus ====================

type mockEventbus struct{}

func (m *mockEventbus) Close() error                                                      { return nil }
func (m *mockEventbus) Publish(ctx context.Context, topic string, message any) error      { return nil }
func (m *mockEventbus) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	return nil
}
func (m *mockEventbus) Unsubscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	return nil
}

// ==================== Mock TaskPool ====================

type mockTaskPool struct{}

func (m *mockTaskPool) AddTask(task func()) error { return nil }
func (m *mockTaskPool) Release()                  {}

// ==================== Mock Cache ====================

type mockCache struct{}

func (m *mockCache) Has(ctx context.Context, key string) (bool, error)                                  { return false, nil }
func (m *mockCache) Get(ctx context.Context, key string, def ...any) cache.Result                       { return nil }
func (m *mockCache) Set(ctx context.Context, key string, value any, expiration ...time.Duration) error  { return nil }
func (m *mockCache) GetSet(ctx context.Context, key string, fn cache.SetValueFunc) cache.Result         { return nil }
func (m *mockCache) Delete(ctx context.Context, keys ...string) (int64, error)                          { return 0, nil }
func (m *mockCache) IncrInt(ctx context.Context, key string, value int64) (int64, error)                { return 0, nil }
func (m *mockCache) IncrFloat(ctx context.Context, key string, value float64) (float64, error)          { return 0, nil }
func (m *mockCache) DecrInt(ctx context.Context, key string, value int64) (int64, error)                { return 0, nil }
func (m *mockCache) DecrFloat(ctx context.Context, key string, value float64) (float64, error)          { return 0, nil }
func (m *mockCache) AddPrefix(key string) string                                                        { return key }
func (m *mockCache) Client() any                                                                        { return nil }
func (m *mockCache) Close() error                                                                       { return nil }

// ==================== Mock LockMaker ====================

type mockLockMaker struct{}

func (m *mockLockMaker) Make(name string) lock.Locker { return nil }
func (m *mockLockMaker) Close() error                 { return nil }

// ==================== Tests ====================

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
	eb := &mockEventbus{}

	ctx.SetEventbus(eb)
	if ctx.Eventbus() != eb {
		t.Fatal("Eventbus not set correctly")
	}

	// 测试设置 nil
	ctx.SetEventbus(nil)
	if ctx.Eventbus() != eb {
		t.Fatal("Eventbus should not change when setting nil")
	}

	// 测试替换
	eb2 := &mockEventbus{}
	ctx.SetEventbus(eb2)
	if ctx.Eventbus() != eb2 {
		t.Fatal("Eventbus not replaced correctly")
	}
}

func TestContext_SetTaskPool(t *testing.T) {
	ctx := NewContext()
	pool := &mockTaskPool{}

	ctx.SetTaskPool(pool)
	if ctx.TaskPool() != pool {
		t.Fatal("TaskPool not set correctly")
	}

	// 测试设置 nil
	ctx.SetTaskPool(nil)
	if ctx.TaskPool() != pool {
		t.Fatal("TaskPool should not change when setting nil")
	}

	// 测试替换
	pool2 := &mockTaskPool{}
	ctx.SetTaskPool(pool2)
	if ctx.TaskPool() != pool2 {
		t.Fatal("TaskPool not replaced correctly")
	}
}

func TestContext_SetCache(t *testing.T) {
	ctx := NewContext()
	ca := &mockCache{}

	ctx.SetCache(ca)
	if ctx.Cache() != ca {
		t.Fatal("Cache not set correctly")
	}

	// 测试设置 nil
	ctx.SetCache(nil)
	if ctx.Cache() != ca {
		t.Fatal("Cache should not change when setting nil")
	}

	// 测试替换
	ca2 := &mockCache{}
	ctx.SetCache(ca2)
	if ctx.Cache() != ca2 {
		t.Fatal("Cache not replaced correctly")
	}
}

func TestContext_SetLockMaker(t *testing.T) {
	ctx := NewContext()
	maker := &mockLockMaker{}

	ctx.SetLockMaker(maker)
	if ctx.LockMaker() != maker {
		t.Fatal("LockMaker not set correctly")
	}

	// 测试设置 nil
	ctx.SetLockMaker(nil)
	if ctx.LockMaker() != maker {
		t.Fatal("LockMaker should not change when setting nil")
	}

	// 测试替换
	maker2 := &mockLockMaker{}
	ctx.SetLockMaker(maker2)
	if ctx.LockMaker() != maker2 {
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

func TestContext_AttachToPackages(t *testing.T) {
	ctx := NewContext()

	// 设置 Logger
	logger := &mockLogger{}
	ctx.SetLogger(logger)

	// 关联到各包
	ctx.AttachToPackages()

	if !ctx.IsAttached() {
		t.Fatal("Context should be attached after AttachToPackages()")
	}

	// 通过 log 包获取应该能获取到 Context 中的 Logger
	if dawnlog.GetLogger() != logger {
		t.Fatal("log.GetLogger() should return the Logger from Context")
	}

	// 解除关联
	ctx.DetachFromPackages()

	if ctx.IsAttached() {
		t.Fatal("Context should not be attached after DetachFromPackages()")
	}

	// 清理
	ctx.Close()
}

func TestContext_AttachToPackages_WithTaskPool(t *testing.T) {
	ctx := NewContext()

	// 设置 TaskPool
	pool := &mockTaskPool{}
	ctx.SetTaskPool(pool)

	// 关联到各包
	ctx.AttachToPackages()

	// 通过 task 包获取应该能获取到 Context 中的 TaskPool
	if task.GetPool() != pool {
		t.Fatal("task.GetPool() should return the TaskPool from Context")
	}

	// 解除关联
	ctx.DetachFromPackages()

	// 清理
	ctx.Close()
}
