package discovery

import (
	"context"
	"net/url"
	"sync"
	"time"

	"github.com/dawnsgo/dawn/cluster"
	"github.com/dawnsgo/dawn/core/endpoint"
	"github.com/dawnsgo/dawn/errors"
	"github.com/dawnsgo/dawn/log"
	"github.com/dawnsgo/dawn/registry"
	cli "github.com/smallnest/rpcx/client"
)

const scheme = "discovery"

const defaultTimeout = 10 * time.Second

type Builder struct {
	dis       registry.Discovery
	ctx       context.Context
	cancel    context.CancelFunc
	watcher   registry.Watcher
	rw        sync.RWMutex
	pairs     map[string][]*cli.KVPair
	resolvers sync.Map
}

func NewBuilder(dis registry.Discovery) (*Builder, error) {
	b := &Builder{}
	b.dis = dis
	b.ctx, b.cancel = context.WithCancel(context.Background())

	if err := b.init(); err != nil {
		return nil, errors.Wrap(err, "init client builder failed")
	}

	return b, nil
}

// MustNewBuilder 创建一个新的 Builder，失败时 panic
// 这是一个便捷函数，适用于初始化阶段，错误表示程序配置有误
func MustNewBuilder(dis registry.Discovery) *Builder {
	b, err := NewBuilder(dis)
	if err != nil {
		panic("dawn/transport/rpcx/resolver/discovery: init client builder failed: " + err.Error())
	}
	return b
}

func (b *Builder) Scheme() string {
	return scheme
}

func (b *Builder) Build(target *url.URL) (cli.ServiceDiscovery, error) {
	b.rw.RLock()
	pairs, ok := b.pairs[target.Host]
	b.rw.RUnlock()

	if !ok {
		return nil, errors.ErrNotFoundServiceAddress
	}

	r := newResolver(target.Host, b)
	r.updateState(pairs)

	b.resolvers.Store(target.Host, r)

	return r, nil
}

func (b *Builder) init() error {
	if b.dis == nil {
		return errors.ErrMissingDiscovery
	}

	ctx, cancel := context.WithTimeout(b.ctx, defaultTimeout)
	watcher, err := b.dis.Watch(ctx, cluster.Mesh.String())
	cancel()
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(b.ctx, defaultTimeout)
	instances, err := b.dis.Services(ctx, cluster.Mesh.String())
	cancel()
	if err != nil {
		_ = watcher.Stop()

		return err
	}

	b.watcher = watcher
	b.updateInstances(instances)

	go b.watch()

	return nil
}

func (b *Builder) watch() {
	for {
		select {
		case <-b.ctx.Done():
			return
		default:
			// exec watch
		}
		instances, err := b.watcher.Next()
		if err != nil {
			continue
		}

		b.updateInstances(instances)
	}
}

func (b *Builder) updateInstances(instances []*registry.ServiceInstance) {
	pairs := make(map[string][]*cli.KVPair, len(instances))
	for _, instance := range instances {
		ep, err := endpoint.ParseEndpoint(instance.Endpoint)
		if err != nil {
			log.Errorf("parse discovery endpoint failed: %v", err)
			continue
		}

		for _, service := range instance.Services {
			pairs[service] = append(pairs[service], &cli.KVPair{Key: "tcp@" + ep.Address()})
		}
	}

	b.rw.Lock()
	b.pairs = pairs
	b.rw.Unlock()

	b.resolvers.Range(func(_, value any) bool {
		r := value.(*Resolver)
		r.updateState(pairs[r.name])

		return true
	})
}

func (b *Builder) removeResolver(r *Resolver) {
	b.resolvers.Delete(r.name)
}
