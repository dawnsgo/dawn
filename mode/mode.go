package mode

import (
	"github.com/dawnsgo/dawn/env"
	"github.com/dawnsgo/dawn/etc"
	"github.com/dawnsgo/dawn/flag"
)

const (
	dawnModeEtcName = "etc.mode"
	dawnModeArgName = "mode"
	dawnModeEnvName = "DAWN_MODE"
)

const (
	// DebugMode indicates dawn mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates dawn mode is release.
	ReleaseMode = "release"
	// TestMode indicates dawn mode is test.
	TestMode = "test"
)

var dawnMode string

// 优先级： 配置文件 < 环境变量 < 运行参数 < mode.SetMode()
func init() {
	mode := etc.Get(dawnModeEtcName, DebugMode).String()
	mode = env.Get(dawnModeEnvName, mode).String()
	mode = flag.String(dawnModeArgName, mode)
	SetMode(mode)
}

// SetMode 设置运行模式
func SetMode(m string) {
	if m == "" {
		m = DebugMode
	}

	switch m {
	case DebugMode, TestMode, ReleaseMode:
		dawnMode = m
	default:
		panic("dawn mode unknown: " + m + " (available mode: debug test release)")
	}
}

// GetMode 获取运行模式
func GetMode() string {
	return dawnMode
}

// IsDebugMode 是否Debug模式
func IsDebugMode() bool {
	return dawnMode == DebugMode
}

// IsTestMode 是否Test模式
func IsTestMode() bool {
	return dawnMode == TestMode
}

// IsReleaseMode 是否Release模式
func IsReleaseMode() bool {
	return dawnMode == ReleaseMode
}
