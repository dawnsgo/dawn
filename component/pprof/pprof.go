package pprof

import (
	"fmt"
	"github.com/dawnsgo/dawn/component"
	"github.com/dawnsgo/dawn/core/info"
	xnet "github.com/dawnsgo/dawn/core/net"
	"github.com/dawnsgo/dawn/log"
	"net/http"
	_ "net/http/pprof"
)

var _ component.Component = &PProf{}

type PProf struct {
	component.Base
	opts *options
}

func NewPProf(opts ...Option) *PProf {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	return &PProf{opts: o}
}

func (*PProf) Name() string {
	return "pprof"
}

func (p *PProf) Start() {
	listenAddr, exposeAddr, err := xnet.ParseAddr(p.opts.addr)
	if err != nil {
		log.Fatalf("pprof addr parse failed: %v", err)
	}

	go func() {
		if err := http.ListenAndServe(listenAddr, nil); err != nil {
			log.Fatalf("pprof server start failed: %v", err)
		}
	}()

	info.PrintBoxInfo("PProf",
		fmt.Sprintf("Url: http://%s/debug/pprof/", exposeAddr),
	)
}
