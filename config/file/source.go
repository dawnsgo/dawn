package file

import (
	"github.com/dawnsgo/dawn/codes"
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/config/file/core"
	"github.com/dawnsgo/dawn/errors"
)

const Name = core.Name

type Source struct {
	opts *options
}

// NewSource 创建一个文件配置源
func NewSource(opts ...Option) (config.Source, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.path == "" {
		return nil, errors.NewWithCode(codes.InvalidConfig, "no config file path specified")
	}

	return core.NewSource(o.path, o.mode), nil
}

// MustNewSource 创建一个文件配置源，失败时 panic
// 这是一个便捷函数，适用于初始化阶段，错误表示程序配置有误
func MustNewSource(opts ...Option) config.Source {
	s, err := NewSource(opts...)
	if err != nil {
		panic("dawn/config/file: create file source failed: " + err.Error())
	}
	return s
}
