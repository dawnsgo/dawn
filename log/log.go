package log

import "sync"

// ContextProvider Context 提供者接口，用于避免循环依赖
type ContextProvider interface {
	// Logger 获取日志记录器
	Logger() Logger
}

var (
	globalLogger    Logger
	contextProvider ContextProvider
	providerMu      sync.RWMutex
)

func init() {
	SetLogger(NewLogger())
}

// SetContextProvider 设置 Context 提供者（由 dawn 包调用）
func SetContextProvider(provider ContextProvider) {
	providerMu.Lock()
	defer providerMu.Unlock()
	contextProvider = provider
}

// GetContextProvider 获取 Context 提供者
func GetContextProvider() ContextProvider {
	providerMu.RLock()
	defer providerMu.RUnlock()
	return contextProvider
}

// SetLogger 设置日志记录器
func SetLogger(logger Logger) {
	if logger == nil {
		return
	}

	providerMu.Lock()
	defer providerMu.Unlock()

	if globalLogger != nil {
		globalLogger.Close()
	}

	globalLogger = logger
}

// GetLogger 获取日志记录器
// 优先从 Context 获取，如果没有关联 Context 则使用全局变量
func GetLogger() Logger {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if logger := contextProvider.Logger(); logger != nil {
			return logger
		}
	}
	return globalLogger
}

// getLogger 内部获取日志记录器（不加锁，供其他函数调用）
func getLogger() Logger {
	providerMu.RLock()
	defer providerMu.RUnlock()

	if contextProvider != nil {
		if logger := contextProvider.Logger(); logger != nil {
			return logger
		}
	}
	return globalLogger
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
	providerMu.Lock()
	defer providerMu.Unlock()

	if globalLogger != nil {
		_ = globalLogger.Close()
	}
}
