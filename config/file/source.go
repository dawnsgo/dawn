package file

import (
	"github.com/dawnsgo/dawn/config"
	"github.com/dawnsgo/dawn/config/file/core"
	"github.com/dawnsgo/dawn/log"
)

const Name = core.Name

type Source struct {
	opts *options
}

func NewSource(opts ...Option) config.Source {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	if o.path == "" {
		log.Fatal("no config file path specified")
	}

	return core.NewSource(o.path, o.mode)
}
