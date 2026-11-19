package discovery

import (
	"context"
	"errors"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"sitehub/common/registry"
)

type fakeWatch struct {
	service string
	update  func([]registry.Instance, error)
}

type fakeDiscovery struct {
	mu           sync.Mutex
	params       map[string]*fakeWatch
	retired      []*fakeWatch
	instances    []registry.Instance
	healthy      bool
	selectError  error
	selectHook   func()
	subscribes   int
	unsubscribes int
}

func (f *fakeDiscovery) Watch(ctx context.Context, service string, update func([]registry.Instance, error)) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p := &fakeWatch{service: service, update: update}
	f.mu.Lock()
	if f.params == nil {
		f.params = make(map[string]*fakeWatch)
	}
	f.params[service] = p
	f.subscribes++
	f.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			f.mu.Lock()
			defer f.mu.Unlock()
			if f.params[service] == p {
				delete(f.params, service)
			}
			f.retired = append(f.retired, p)
			f.unsubscribes++
		})
	}, nil
}

func (f *fakeDiscovery) List(ctx context.Context, _ string) ([]registry.Instance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	instances, err, hook := slices.Clone(f.instances), f.selectError, f.selectHook
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
	return instances, err
}

func (f *fakeDiscovery) Healthy() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.healthy
}

func (f *fakeDiscovery) emit(service string, instances []registry.Instance, err error) {
	f.mu.Lock()
	if err == nil {
		f.instances = slices.Clone(instances)
	}
	p := f.params[service]
	f.mu.Unlock()
	if p != nil {
		p.update(instances, err)
	}
}

func (f *fakeDiscovery) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.subscribes, f.unsubscribes
}

func healthyInstance(ip string) registry.Instance {
	return registry.Instance{IP: ip, Port: 8801, Weight: 100, Enabled: true, Healthy: true}
}

func next(t *testing.T, changes <-chan []Node, want ...Node) {
	t.Helper()
	select {
	case got := <-changes:
		if !slices.Equal(got, want) {
			t.Fatalf("nodes = %v, want %v", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for nodes %v", want)
	}
}

func TestWatchSupportsConfiguredHTTPServiceNames(t *testing.T) {
	for _, service := range []string{"orders", "Payments.v2", "legacy_service", "3rd-party", strings.Repeat("s", 128)} {
		t.Run(service, func(t *testing.T) {
			f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
			m := New(f, time.Second)
			defer m.Close()
			changes := make(chan []Node, 8)
			cancel, err := m.Watch(service, func(nodes []Node) { changes <- nodes })
			if err != nil {
				t.Fatalf("configured service rejected: %v", err)
			}
			defer cancel()
			next(t, changes, Node{URL: "http://127.0.0.1:8801", Weight: 100})
			f.mu.Lock()
			param := f.params[service]
			f.mu.Unlock()
			if param == nil || param.service != service {
				t.Fatalf("provider subscription did not preserve configured service: %+v", param)
			}
			f.emit(service, []registry.Instance{healthyInstance("127.0.0.2")}, nil)
			next(t, changes, Node{URL: "http://127.0.0.2:8801", Weight: 100})
		})
	}
}

func TestWatchRejectsInvalidNamesBeforeSubscribing(t *testing.T) {
	f := &fakeDiscovery{healthy: true}
	m := New(f, time.Second)
	defer m.Close()
	for _, service := range []string{"", " orders", "orders ", "orders/api", "orders:8080", "orders?version=1", "orders\n", "-orders", strings.Repeat("s", 129)} {
		if cancel, err := m.Watch(service, func([]Node) {}); err == nil {
			cancel()
			t.Errorf("accepted invalid service name %q", service)
		}
	}
	if sub, unsub := f.counts(); sub != 0 || unsub != 0 {
		t.Fatalf("invalid names reached the discovery provider: subscribe=%d unsubscribe=%d", sub, unsub)
	}
}

func TestWatchSharesSubscriptionAndCancellationDoesNotAffectOtherSnapshot(t *testing.T) {
	f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, time.Second)
	defer m.Close()
	a, b := make(chan []Node, 8), make(chan []Node, 8)
	cancelA, err := m.Watch("blog-api", func(nodes []Node) { a <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	cancelB, err := m.Watch("blog-api", func(nodes []Node) { b <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	node := Node{URL: "http://127.0.0.1:8801", Weight: 100}
	next(t, a, node)
	next(t, b, node)
	if sub, unsub := f.counts(); sub != 1 || unsub != 0 {
		t.Fatalf("provider counts = %d/%d; expected one shared subscription", sub, unsub)
	}
	cancelA()
	cancelA()
	f.emit("blog-api", []registry.Instance{healthyInstance("127.0.0.2")}, nil)
	next(t, b, Node{URL: "http://127.0.0.2:8801", Weight: 100})
	if _, unsub := f.counts(); unsub != 0 {
		t.Fatal("canceling one watcher unsubscribed another snapshot")
	}
	cancelB()
	if sub, unsub := f.counts(); sub != 1 || unsub != 1 {
		t.Fatalf("provider counts = %d/%d after last cancel", sub, unsub)
	}
	// A new generation gets a new subscription after the last one is removed.
	cancelC, err := m.Watch("blog-api", func(nodes []Node) { b <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	next(t, b, Node{URL: "http://127.0.0.2:8801", Weight: 100})
	cancelC()
	if sub, unsub := f.counts(); sub != 2 || unsub != 2 {
		t.Fatalf("provider counts = %d/%d after replacement", sub, unsub)
	}
}

func TestAuthoritativeEmptyAndFilteringRemoveNodesImmediately(t *testing.T) {
	f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, time.Second)
	defer m.Close()
	changes := make(chan []Node, 8)
	_, err := m.Watch("file-api", func(nodes []Node) { changes <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	next(t, changes, Node{URL: "http://127.0.0.1:8801", Weight: 100})
	f.emit("file-api", nil, nil)
	next(t, changes)
	valid := healthyInstance("2001:db8::1")
	f.emit("file-api", []registry.Instance{valid}, nil)
	next(t, changes, Node{URL: "http://[2001:db8::1]:8801", Weight: 100})
	valid.Enabled = false
	f.emit("file-api", []registry.Instance{valid}, nil)
	next(t, changes)

	bad := []registry.Instance{}
	for _, ip := range []string{"0.0.0.0", "::", "224.0.0.1", "ff02::1", "blog-api", "127.0.0.1/path", "127.0.0.1:80"} {
		bad = append(bad, healthyInstance(ip))
	}
	for _, port := range []uint64{0, 65536} {
		i := healthyInstance("127.0.0.1")
		i.Port = port
		bad = append(bad, i)
	}
	for _, weight := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		i := healthyInstance("127.0.0.1")
		i.Weight = weight
		bad = append(bad, i)
	}
	i := healthyInstance("127.0.0.1")
	i.Healthy = false
	bad = append(bad, i)
	if nodes := nodesFromInstances(bad); len(nodes) != 0 {
		t.Fatalf("invalid registry endpoints accepted: %v", nodes)
	}
}

func TestDiscoveryFailuresRetainLastGoodUntilTTLAndRecover(t *testing.T) {
	for _, disconnect := range []bool{false, true} {
		t.Run(map[bool]string{false: "read failure", true: "disconnected"}[disconnect], func(t *testing.T) {
			instance := healthyInstance("127.0.0.1")
			f := &fakeDiscovery{healthy: true, instances: []registry.Instance{instance}}
			m := New(f, 100*time.Millisecond)
			defer m.Close()
			changes := make(chan []Node, 8)
			_, err := m.Watch("blog-api", func(nodes []Node) { changes <- nodes })
			if err != nil {
				t.Fatal(err)
			}
			node := Node{URL: "http://127.0.0.1:8801", Weight: 100}
			next(t, changes, node)
			f.mu.Lock()
			f.healthy = !disconnect
			f.selectError = errors.New("registry unavailable")
			f.mu.Unlock()
			f.emit("blog-api", nil, errors.New("push transport failed"))
			select {
			case <-changes:
				t.Fatal("transient error immediately cleared the last good snapshot")
			case <-time.After(20 * time.Millisecond):
			}
			next(t, changes)
			f.mu.Lock()
			f.healthy, f.selectError = true, nil
			f.mu.Unlock()
			next(t, changes, node)
		})
	}
}

func TestInitialSelectionCannotOverwriteNewerPush(t *testing.T) {
	f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	f.selectHook = func() { once.Do(func() { close(started); <-release }) }
	m := New(f, time.Second)
	defer m.Close()
	changes := make(chan []Node, 8)
	watchDone := make(chan error, 1)
	go func() {
		_, err := m.Watch("blog-api", func(nodes []Node) { changes <- nodes })
		watchDone <- err
	}()
	<-started
	f.emit("blog-api", []registry.Instance{healthyInstance("127.0.0.2")}, nil)
	close(release)
	if err := <-watchDone; err != nil {
		t.Fatal(err)
	}
	next(t, changes, Node{URL: "http://127.0.0.2:8801", Weight: 100})
}

func TestSlowProviderReadCannotPostponeStaleExpiry(t *testing.T) {
	f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, 100*time.Millisecond)
	defer m.Close()
	changes := make(chan []Node, 8)
	_, err := m.Watch("blog-api", func(nodes []Node) { changes <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	next(t, changes, Node{URL: "http://127.0.0.1:8801", Weight: 100})
	started, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	var once sync.Once
	f.mu.Lock()
	f.selectHook = func() { once.Do(func() { close(started); <-release }) }
	f.mu.Unlock()
	<-started
	// This update must arrive before the deliberately stalled provider read returns.
	next(t, changes)
}

func TestSlowCallbackDoesNotBlockOtherWatchersOrProviderPush(t *testing.T) {
	f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, time.Second)
	defer m.Close()
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	_, err := m.Watch("blog-api", func([]Node) { once.Do(func() { close(started); <-release }) })
	if err != nil {
		t.Fatal(err)
	}
	<-started
	defer close(release)
	changes := make(chan []Node, 8)
	_, err = m.Watch("blog-api", func(nodes []Node) { changes <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	next(t, changes, Node{URL: "http://127.0.0.1:8801", Weight: 100})
	for n := 0; n < 10; n++ {
		f.emit("blog-api", []registry.Instance{healthyInstance("127.0.0.2")}, nil)
	}
	next(t, changes, Node{URL: "http://127.0.0.2:8801", Weight: 100})
}

func TestCallbackCanCloseManagerAndLatePushCannotReviveSubscription(t *testing.T) {
	f := &fakeDiscovery{healthy: true, instances: []registry.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, time.Second)
	done := make(chan struct{})
	_, err := m.Watch("blog-api", func([]Node) { m.Close(); close(done) })
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close deadlocked inside update callback")
	}
	m.Close()
	if sub, unsub := f.counts(); sub != 1 || unsub != 1 {
		t.Fatalf("provider counts = %d/%d after Close", sub, unsub)
	}
	f.mu.Lock()
	retired := f.retired[0]
	f.mu.Unlock()
	retired.update([]registry.Instance{healthyInstance("127.0.0.2")}, nil)
	if _, err := m.Watch("blog-api", func([]Node) {}); err == nil {
		t.Fatal("closed manager accepted another subscription")
	}
}

func TestNodesFromInstancesPreservesWeightsAndDeduplicates(t *testing.T) {
	fractional := healthyInstance("127.0.0.1")
	fractional.Weight = 0.25
	duplicate := fractional
	duplicate.Weight = 2.6
	large := healthyInstance("127.0.0.2")
	large.Weight = math.MaxFloat64
	want := []Node{{URL: "http://127.0.0.1:8801", Weight: 3}, {URL: "http://127.0.0.2:8801", Weight: 1_000_000}}
	if got := nodesFromInstances([]registry.Instance{large, fractional, duplicate}); !slices.Equal(got, want) {
		t.Fatalf("nodes = %v, want %v", got, want)
	}
}

type openingDiscovery struct{ started chan struct{} }

func (d openingDiscovery) Watch(ctx context.Context, _ string, _ func([]registry.Instance, error)) (func(), error) {
	close(d.started)
	<-ctx.Done()
	return nil, ctx.Err()
}
func (openingDiscovery) List(context.Context, string) ([]registry.Instance, error) { return nil, nil }
func (openingDiscovery) Healthy() bool                                             { return true }

func TestCloseCancelsOpeningProviderWatch(t *testing.T) {
	d := openingDiscovery{started: make(chan struct{})}
	m := New(d, time.Second)
	watchDone := make(chan error, 1)
	go func() {
		_, err := m.Watch("blog-api", func([]Node) {})
		watchDone <- err
	}()
	<-d.started
	closeDone := make(chan struct{})
	go func() { m.Close(); close(closeDone) }()
	select {
	case <-closeDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not cancel the opening provider watch")
	}
	select {
	case err := <-watchDone:
		if err == nil {
			t.Fatal("opening watch succeeded after manager Close")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("opening Watch did not return after cancellation")
	}
}
