package discovery

import (
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type fakeNaming struct {
	mu           sync.Mutex
	params       map[string]*vo.SubscribeParam
	retired      []*vo.SubscribeParam
	instances    []model.Instance
	healthy      bool
	selectError  error
	selectHook   func()
	subscribes   int
	unsubscribes int
}

func (f *fakeNaming) Subscribe(p *vo.SubscribeParam) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.params == nil {
		f.params = make(map[string]*vo.SubscribeParam)
	}
	f.params[p.ServiceName] = p
	f.subscribes++
	return nil
}

func (f *fakeNaming) Unsubscribe(p *vo.SubscribeParam) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.params, p.ServiceName)
	f.retired = append(f.retired, p)
	f.unsubscribes++
	return nil
}

func (f *fakeNaming) SelectInstances(vo.SelectInstancesParam) ([]model.Instance, error) {
	f.mu.Lock()
	instances, err, hook := slices.Clone(f.instances), f.selectError, f.selectHook
	f.mu.Unlock()
	if hook != nil {
		hook()
	}
	return instances, err
}

func (f *fakeNaming) ServerHealthy() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.healthy
}

func (f *fakeNaming) emit(service string, instances []model.Instance, err error) {
	f.mu.Lock()
	if err == nil {
		f.instances = slices.Clone(instances)
	}
	p := f.params[service]
	f.mu.Unlock()
	if p != nil {
		p.SubscribeCallback(instances, err)
	}
}

func (f *fakeNaming) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.subscribes, f.unsubscribes
}

func healthyInstance(ip string) model.Instance {
	return model.Instance{Ip: ip, Port: 8801, Weight: 100, Enable: true, Healthy: true}
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

func TestWatchSharesSubscriptionAndCancellationDoesNotAffectOtherSnapshot(t *testing.T) {
	f := &fakeNaming{healthy: true, instances: []model.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, "SITEHUB", time.Second)
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
		t.Fatalf("SDK counts = %d/%d; expected one shared subscription", sub, unsub)
	}
	cancelA()
	cancelA()
	f.emit("blog-api", []model.Instance{healthyInstance("127.0.0.2")}, nil)
	next(t, b, Node{URL: "http://127.0.0.2:8801", Weight: 100})
	if _, unsub := f.counts(); unsub != 0 {
		t.Fatal("canceling one watcher unsubscribed another snapshot")
	}
	cancelB()
	if sub, unsub := f.counts(); sub != 1 || unsub != 1 {
		t.Fatalf("SDK counts = %d/%d after last cancel", sub, unsub)
	}
	// A new generation gets a new subscription after the last one is removed.
	cancelC, err := m.Watch("blog-api", func(nodes []Node) { b <- nodes })
	if err != nil {
		t.Fatal(err)
	}
	next(t, b, Node{URL: "http://127.0.0.2:8801", Weight: 100})
	cancelC()
	if sub, unsub := f.counts(); sub != 2 || unsub != 2 {
		t.Fatalf("SDK counts = %d/%d after replacement", sub, unsub)
	}
}

func TestAuthoritativeEmptyAndFilteringRemoveNodesImmediately(t *testing.T) {
	f := &fakeNaming{healthy: true, instances: []model.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, "SITEHUB", time.Second)
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
	f.emit("file-api", []model.Instance{valid}, nil)
	next(t, changes, Node{URL: "http://[2001:db8::1]:8801", Weight: 100})
	valid.Enable = false
	f.emit("file-api", []model.Instance{valid}, nil)
	next(t, changes)

	bad := []model.Instance{}
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
			f := &fakeNaming{healthy: true, instances: []model.Instance{instance}}
			m := New(f, "SITEHUB", 100*time.Millisecond)
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
	f := &fakeNaming{healthy: true, instances: []model.Instance{healthyInstance("127.0.0.1")}}
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	f.selectHook = func() { once.Do(func() { close(started); <-release }) }
	m := New(f, "SITEHUB", time.Second)
	defer m.Close()
	changes := make(chan []Node, 8)
	watchDone := make(chan error, 1)
	go func() {
		_, err := m.Watch("blog-api", func(nodes []Node) { changes <- nodes })
		watchDone <- err
	}()
	<-started
	f.emit("blog-api", []model.Instance{healthyInstance("127.0.0.2")}, nil)
	close(release)
	if err := <-watchDone; err != nil {
		t.Fatal(err)
	}
	next(t, changes, Node{URL: "http://127.0.0.2:8801", Weight: 100})
}

func TestSlowSDKReadCannotPostponeStaleExpiry(t *testing.T) {
	f := &fakeNaming{healthy: true, instances: []model.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, "SITEHUB", 100*time.Millisecond)
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
	// This update must arrive before the deliberately stalled SDK read returns.
	next(t, changes)
}

func TestSlowCallbackDoesNotBlockOtherWatchersOrSDKPush(t *testing.T) {
	f := &fakeNaming{healthy: true, instances: []model.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, "SITEHUB", time.Second)
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
		f.emit("blog-api", []model.Instance{healthyInstance("127.0.0.2")}, nil)
	}
	next(t, changes, Node{URL: "http://127.0.0.2:8801", Weight: 100})
}

func TestCallbackCanCloseManagerAndLatePushCannotReviveSubscription(t *testing.T) {
	f := &fakeNaming{healthy: true, instances: []model.Instance{healthyInstance("127.0.0.1")}}
	m := New(f, "SITEHUB", time.Second)
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
		t.Fatalf("SDK counts = %d/%d after Close", sub, unsub)
	}
	f.mu.Lock()
	retired := f.retired[0]
	f.mu.Unlock()
	retired.SubscribeCallback([]model.Instance{healthyInstance("127.0.0.2")}, nil)
	if _, err := m.Watch("blog-api", func([]Node) {}); err == nil {
		t.Fatal("closed manager accepted another subscription")
	}
}
