package registry

import (
	"context"
	"fmt"
	"math"
	"net"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc/resolver"
)

// Builder translates the stable discovery:/// target scheme into addresses
// through the injected proxy, regardless of its configured provider.
type Builder struct {
	Naming        Discovery
	MaxStale      time.Duration
	mu            sync.Mutex
	closed        bool
	subscriptions map[*subscription]struct{}
}

func (b *Builder) Scheme() string { return "discovery" }

var serviceName = regexp.MustCompile(`^[a-z][a-z0-9-]*-rpc$`)

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	name := target.Endpoint()
	if target.URL.Host != "" || target.URL.RawQuery != "" || !serviceName.MatchString(name) || b.Naming == nil || b.MaxStale <= 0 {
		return nil, fmt.Errorf("RPC target must be discovery:///service-rpc with discovery and positive MaxStale")
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &subscription{builder: b, cc: cc, service: name, ctx: ctx, cancel: cancel, lastGood: time.Now()}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		cancel()
		return nil, fmt.Errorf("registry resolver is closed")
	}
	if b.subscriptions == nil {
		b.subscriptions = make(map[*subscription]struct{})
	}
	b.subscriptions[r] = struct{}{}
	b.mu.Unlock()
	stop, err := b.Naming.Watch(ctx, name, r.update)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("subscribe to %s failed", name)
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		stop()
		return nil, fmt.Errorf("registry resolver is closed")
	}
	r.stop = stop
	r.mu.Unlock()
	go r.monitor()
	return r, nil
}

func (b *Builder) Close() {
	b.mu.Lock()
	b.closed = true
	all := make([]*subscription, 0, len(b.subscriptions))
	for r := range b.subscriptions {
		all = append(all, r)
	}
	b.mu.Unlock()
	for _, r := range all {
		r.Close()
	}
}

type subscription struct {
	builder         *Builder
	cc              resolver.ClientConn
	service         string
	ctx             context.Context
	cancel          context.CancelFunc
	stop            func()
	mu              sync.Mutex
	closed, expired bool
	revision        uint64
	lastGood        time.Time
	lastAddresses   []string
	hasSnapshot     bool
}

func (r *subscription) update(instances []Instance, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	r.revision++
	if err != nil {
		r.cc.ReportError(fmt.Errorf("service discovery unavailable"))
		return
	}
	r.apply(instances)
}

func (r *subscription) apply(instances []Instance) {
	r.lastGood = time.Now()
	r.expired = false
	addresses := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		ip := net.ParseIP(instance.IP)
		if !instance.Healthy || !instance.Enabled || instance.Weight <= 0 || math.IsNaN(instance.Weight) || math.IsInf(instance.Weight, 0) || instance.Port == 0 || instance.Port > 65535 || ip == nil || ip.IsUnspecified() || ip.IsMulticast() {
			continue
		}
		addresses[net.JoinHostPort(ip.String(), strconv.FormatUint(instance.Port, 10))] = struct{}{}
	}
	keys := make([]string, 0, len(addresses))
	for address := range addresses {
		keys = append(keys, address)
	}
	sort.Strings(keys)
	if r.hasSnapshot && slices.Equal(r.lastAddresses, keys) {
		return
	}
	r.hasSnapshot, r.lastAddresses = true, keys
	state := resolver.State{Addresses: make([]resolver.Address, 0, len(keys))}
	for _, key := range keys {
		state.Addresses = append(state.Addresses, resolver.Address{Addr: key})
	}
	_ = r.cc.UpdateState(state)
}

func (r *subscription) refresh() {
	if !r.builder.Naming.Healthy() {
		return
	}
	r.mu.Lock()
	revision, closed := r.revision, r.closed
	r.mu.Unlock()
	if closed {
		return
	}
	instances, err := r.builder.Naming.List(r.ctx, r.service)
	r.mu.Lock()
	defer r.mu.Unlock()
	// A later watch event wins over an older in-flight list, including errors.
	if r.closed || revision != r.revision || err != nil {
		return
	}
	r.apply(instances)
}

func (r *subscription) monitor() {
	refresh := make(chan struct{}, 1)
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-refresh:
				r.refresh()
			}
		}
	}()
	defer func() { <-workerDone }()
	ticker := time.NewTicker(max(time.Nanosecond, min(time.Second, r.builder.MaxStale/4)))
	defer ticker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case now := <-ticker.C:
			select {
			case refresh <- struct{}{}:
			default:
			}
			r.mu.Lock()
			if !r.closed && !r.expired && now.Sub(r.lastGood) >= r.builder.MaxStale {
				r.expired = true
				r.hasSnapshot, r.lastAddresses = true, nil
				_ = r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{}})
				r.cc.ReportError(fmt.Errorf("service discovery exceeded MaxStale"))
			}
			r.mu.Unlock()
		}
	}
}

func (r *subscription) ResolveNow(resolver.ResolveNowOptions) {}
func (r *subscription) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	r.cancel()
	stop := r.stop
	r.mu.Unlock()
	if stop != nil {
		stop()
	}
	r.builder.mu.Lock()
	delete(r.builder.subscriptions, r)
	r.builder.mu.Unlock()
}
