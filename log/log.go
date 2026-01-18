package log

import (
	"sync"
	"sync/atomic"
)

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// Logger 获取日志记录器
	Logger() Logger
}

var (
	// globalLogger 使用 atomic.Value 存储，无锁读取
	globalLogger atomic.Value // Logger

	// contextProvider 使用 atomic.Value 存储，无锁读取
	contextProvider atomic.Value // ContextProvider

	// writeMu 保护写入操作的原子性
	writeMu sync.Mutex
)

func init() {
	SetLogger(NewLogger())
}

// SetContextProvider 设置 Context 提供者（由 dawn 包调用）
func SetContextProvider(provider ContextProvider) {
	if provider == nil {
		contextProvider.Store((*contextProviderWrapper)(nil))
	} else {
		contextProvider.Store(&contextProviderWrapper{provider})
	}
}

// contextProviderWrapper 包装器，解决 atomic.Value 存储 nil 接口的问题
type contextProviderWrapper struct {
	provider ContextProvider
}

// GetContextProvider 获取 Context 提供者
func GetContextProvider() ContextProvider {
	if v := contextProvider.Load(); v != nil {
		if w, ok := v.(*contextProviderWrapper); ok && w != nil {
			return w.provider
		}
	}
	return nil
}

// SetLogger 设置日志记录器
func SetLogger(logger Logger) {
	if logger == nil {
		return
	}

	writeMu.Lock()
	defer writeMu.Unlock()

	// 关闭旧的 logger
	if old := getGlobalLogger(); old != nil {
		old.Close()
	}

	globalLogger.Store(logger)
}

// getGlobalLogger 获取全局 logger（无锁）
func getGlobalLogger() Logger {
	if v := globalLogger.Load(); v != nil {
		return v.(Logger)
	}
	return nil
}

// GetLogger 获取日志记录器
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetLogger() Logger {
	if provider := GetContextProvider(); provider != nil {
		if logger := provider.Logger(); logger != nil {
			return logger
		}
	}
	return getGlobalLogger()
}

// getLogger 内部获取日志记录器（供其他函数调用）
func getLogger() Logger {
	return GetLogger()
}

// Print 打印日志，不含堆栈信息
func Print(level Level, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Print(level, a...)
	}
}

// Printf 打印模板日志，不含堆栈信息
func Printf(level Level, format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Printf(level, format, a...)
	}
}

// Debug 打印调试日志
func Debug(a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Debug(a...)
	}
}

// Debugf 打印调试模板日志
func Debugf(format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Debugf(format, a...)
	}
}

// Info 打印信息日志
func Info(a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Info(a...)
	}
}

// Infof 打印信息模板日志
func Infof(format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Infof(format, a...)
	}
}

// Warn 打印警告日志
func Warn(a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Warn(a...)
	}
}

// Warnf 打印警告模板日志
func Warnf(format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Warnf(format, a...)
	}
}

// Error 打印错误日志
func Error(a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Error(a...)
	}
}

// Errorf 打印错误模板日志
func Errorf(format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Errorf(format, a...)
	}
}

// Fatal 打印致命错误日志
func Fatal(a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Fatal(a...)
	}
}

// Fatalf 打印致命错误模板日志
func Fatalf(format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Fatalf(format, a...)
	}
}

// Panic 打印Panic日志
func Panic(a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Panic(a...)
	}
}

// Panicf 打印Panic模板日志
func Panicf(format string, a ...any) {
	if logger := getLogger(); logger != nil {
		logger.Panicf(format, a...)
	}
}

// Close 关闭日志
func Close() {
	writeMu.Lock()
	defer writeMu.Unlock()

	if logger := getGlobalLogger(); logger != nil {
		_ = logger.Close()
	}
}
