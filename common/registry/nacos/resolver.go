package nacos

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc/resolver"
)

type Naming interface {
	Subscribe(*vo.SubscribeParam) error
	Unsubscribe(*vo.SubscribeParam) error
	SelectInstances(vo.SelectInstancesParam) ([]model.Instance, error)
	ServerHealthy() bool
}
type Builder struct {
	Naming        Naming
	Group         string
	MaxStale      time.Duration
	mu            sync.Mutex
	closed        bool
	subscriptions map[*subscription]struct{}
}

func (b *Builder) Scheme() string { return "nacos" }

var serviceName = regexp.MustCompile(`^[a-z][a-z0-9-]*-rpc$`)

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, fmt.Errorf("registry resolver is closed")
	}
	name := target.Endpoint()
	if target.URL.Host != "" || target.URL.RawQuery != "" || !serviceName.MatchString(name) || b.MaxStale <= 0 {
		return nil, fmt.Errorf("RPC target must be nacos:///service-rpc and MaxStale must be positive")
	}
	r := &subscription{builder: b, cc: cc, done: make(chan struct{}), lastHealthy: time.Now()}
	r.param = &vo.SubscribeParam{ServiceName: name, GroupName: b.Group, SubscribeCallback: r.update}
	if err := b.Naming.Subscribe(r.param); err != nil {
		return nil, fmt.Errorf("subscribe to %s failed", name)
	}
	// Subscribe installs the callback first, so initial resolution cannot lose a
	// later update. Both the SDK cache and callback always retain valid empties.
	r.refresh()
	go r.monitor()
	if b.subscriptions == nil {
		b.subscriptions = make(map[*subscription]struct{})
	}
	b.subscriptions[r] = struct{}{}
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
	builder     *Builder
	cc          resolver.ClientConn
	param       *vo.SubscribeParam
	mu          sync.Mutex
	closed      bool
	expired     bool
	lastHealthy time.Time
	done        chan struct{}
}

func (r *subscription) update(instances []model.Instance, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	if err != nil {
		r.cc.ReportError(fmt.Errorf("Nacos discovery unavailable"))
		return
	}
	r.lastHealthy = time.Now()
	r.expired = false
	addresses := make([]resolver.Address, 0, len(instances))
	for _, i := range instances {
		if i.Healthy && i.Enable && i.Weight > 0 && i.Port > 0 {
			addresses = append(addresses, resolver.Address{Addr: net.JoinHostPort(i.Ip, strconv.FormatUint(i.Port, 10))})
		}
	}
	// An authoritative empty list MUST remove departed endpoints immediately.
	_ = r.cc.UpdateState(resolver.State{Addresses: addresses})
}
func (r *subscription) refresh() {
	if !r.builder.Naming.ServerHealthy() {
		return
	}
	instances, err := r.builder.Naming.SelectInstances(vo.SelectInstancesParam{ServiceName: r.param.ServiceName, GroupName: r.param.GroupName, HealthyOnly: true})
	r.update(instances, err)
}
func (r *subscription) monitor() {
	ticker := time.NewTicker(min(time.Second, r.builder.MaxStale/2))
	defer ticker.Stop()
	for {
		select {
		case <-r.done:
			return
		case now := <-ticker.C:
			healthy := r.builder.Naming.ServerHealthy()
			r.mu.Lock()
			if r.closed {
				r.mu.Unlock()
				return
			}
			if healthy && !r.expired {
				r.lastHealthy = now
			}
			recovering := healthy && r.expired
			if !healthy && !r.expired && now.Sub(r.lastHealthy) >= r.builder.MaxStale {
				r.expired = true
				_ = r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{}})
				r.cc.ReportError(fmt.Errorf("Nacos disconnected beyond MaxStale"))
			}
			r.mu.Unlock()
			if recovering {
				r.refresh()
			}
		}
	}
}
func (r *subscription) ResolveNow(resolver.ResolveNowOptions) { /* Nacos pushes updates; monitor handles reconnection. */
}
func (r *subscription) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	close(r.done)
	r.mu.Unlock()
	_ = r.builder.Naming.Unsubscribe(r.param)
	r.builder.mu.Lock()
	delete(r.builder.subscriptions, r)
	r.builder.mu.Unlock()
	// The process owns the shared NamingClient, never an individual resolver.
}
