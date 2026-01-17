package file

import (
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/config/file/core"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
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
		return nil, errors.NewError("no config file path specified")
	}

	return core.NewSource(o.path, o.mode), nil
}

// MustNewSource 创建一个文件配置源，失败时 panic
func MustNewSource(opts ...Option) config.Source {
	s, err := NewSource(opts...)
	if err != nil {
		log.Fatalf("create file source failed: %v", err)
	}
	return s
}
