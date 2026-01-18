package dawn

import (
	"github.com/dawnsgo/dawn/cache"
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/eventbus"
	"github.com/dawnsgo/dawn/lock"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/task"
)

// ==================== Setter 函数 ====================
// 这些函数只需要设置到 Context，各包会自动从 Context 读取
// 不再需要双重同步（Context + 全局变量）

// SetLogger 设置日志记录器
func SetLogger(logger log.Logger) {
	Default().SetLogger(logger)
}

// SetConfigurator 设置配置器
func SetConfigurator(configurator config.Configurator) {
	Default().SetConfigurator(configurator)
}

// SetEventbus 设置事件总线
func SetEventbus(eb eventbus.Eventbus) {
	Default().SetEventbus(eb)
}

// SetTaskPool 设置任务池
func SetTaskPool(pool task.Pool) {
	Default().SetTaskPool(pool)
}

// SetCache 设置缓存
func SetCache(ca cache.Cache) {
	Default().SetCache(ca)
}

// SetLockMaker 设置分布式锁制造器
func SetLockMaker(maker lock.Maker) {
	Default().SetLockMaker(maker)
}

// ==================== Getter 函数 ====================
// 这些函数从 Context 读取

// GetLogger 获取日志记录器
func GetLogger() log.Logger {
	return Default().Logger()
}

// GetConfigurator 获取配置器
func GetConfigurator() config.Configurator {
	return Default().Configurator()
}

// GetEventbus 获取事件总线
func GetEventbus() eventbus.Eventbus {
	return Default().Eventbus()
}

// GetTaskPool 获取任务池
func GetTaskPool() task.Pool {
	return Default().TaskPool()
}

// GetCache 获取缓存
func GetCache() cache.Cache {
	return Default().Cache()
}

// GetLockMaker 获取分布式锁制造器
func GetLockMaker() lock.Maker {
	return Default().LockMaker()
}
