package registry

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/resolver"
)

type fakeDiscovery struct {
	mu        sync.Mutex
	update    func([]Instance, error)
	healthy   bool
	instances []Instance
	canceled  bool
	listHook  func()
}

func (f *fakeDiscovery) Watch(_ context.Context, _ string, update func([]Instance, error)) (func(), error) {
	f.mu.Lock()
	f.update = update
	initial := slices.Clone(f.instances)
	f.mu.Unlock()
	update(initial, nil)
	return func() { f.mu.Lock(); f.canceled = true; f.mu.Unlock() }, nil
}
func (f *fakeDiscovery) List(context.Context, string) ([]Instance, error) {
	f.mu.Lock()
	values, hook := slices.Clone(f.instances), f.listHook
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
	return values, nil
}
func (f *fakeDiscovery) Healthy() bool { f.mu.Lock(); defer f.mu.Unlock(); return f.healthy }
func (f *fakeDiscovery) emit(instances []Instance, err error) {
	f.mu.Lock()
	if err == nil {
		f.instances = slices.Clone(instances)
	}
	update := f.update
	f.mu.Unlock()
	update(instances, err)
}

type captureConn struct {
	resolver.ClientConn
	states chan resolver.State
}

func (c *captureConn) UpdateState(s resolver.State) error { c.states <- s; return nil }
func (c *captureConn) ReportError(error)                  {}

func nextCount(t *testing.T, states <-chan resolver.State, want int) {
	t.Helper()
	select {
	case state := <-states:
		if len(state.Addresses) != want {
			t.Fatalf("addresses=%v, want %d", state.Addresses, want)
		}
	case <-time.After(time.Second):
		t.Fatal("discovery update timed out")
	}
}

func TestResolverEmptyOutageRecoveryAndCancel(t *testing.T) {
	instance := Instance{IP: "127.0.0.1", Port: 9801, Healthy: true, Enabled: true, Weight: 100}
	discovery := &fakeDiscovery{healthy: true, instances: []Instance{instance}}
	cc := &captureConn{states: make(chan resolver.State, 20)}
	u, _ := url.Parse("discovery:///rbac-rpc")
	builder := &Builder{Naming: discovery, MaxStale: 40 * time.Millisecond}
	r, err := builder.Build(resolver.Target{URL: *u}, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer builder.Close()
	nextCount(t, cc.states, 1)
	discovery.emit(nil, nil)
	nextCount(t, cc.states, 0)
	discovery.emit([]Instance{instance}, nil)
	nextCount(t, cc.states, 1)
	discovery.mu.Lock()
	discovery.healthy = false
	discovery.mu.Unlock()
	discovery.emit(nil, errors.New("backend unavailable"))
	select {
	case <-cc.states:
		t.Fatal("a transient error immediately removed the last good snapshot")
	default:
	}
	nextCount(t, cc.states, 0)
	discovery.mu.Lock()
	discovery.healthy = true
	discovery.mu.Unlock()
	nextCount(t, cc.states, 1)
	r.Close()
	discovery.mu.Lock()
	canceled := discovery.canceled
	discovery.mu.Unlock()
	if !canceled {
		t.Fatal("watch leaked")
	}
	discovery.emit(nil, nil)
	select {
	case <-cc.states:
		t.Fatal("update after close")
	default:
	}
}

func TestResolverNewerWatchWinsOverInFlightList(t *testing.T) {
	instance := Instance{IP: "127.0.0.1", Port: 9801, Healthy: true, Enabled: true, Weight: 100}
	discovery := &fakeDiscovery{healthy: true, instances: []Instance{instance}}
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	discovery.listHook = func() { once.Do(func() { close(started); <-release }) }
	cc := &captureConn{states: make(chan resolver.State, 20)}
	u, _ := url.Parse("discovery:///rbac-rpc")
	builder := &Builder{Naming: discovery, MaxStale: time.Second}
	_, err := builder.Build(resolver.Target{URL: *u}, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer builder.Close()
	nextCount(t, cc.states, 1)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("list did not start")
	}
	discovery.emit(nil, nil)
	nextCount(t, cc.states, 0)
	close(release)
	select {
	case state := <-cc.states:
		t.Fatalf("old list revived removed endpoint: %+v", state)
	case <-time.After(80 * time.Millisecond):
	}
}

func TestAdvertiseRejectsWildcardAndLoopbackByDefault(t *testing.T) {
	for _, ip := range []string{"0.0.0.0", "127.0.0.1"} {
		if _, _, err := ResolveAdvertise(ip, "0.0.0.0:9801", false); err == nil {
			t.Fatal("unsafe advertise address accepted")
		}
	}
	if _, _, err := ResolveAdvertise("127.0.0.1", "0.0.0.0:9801", true); err != nil {
		t.Fatal(err)
	}
}
