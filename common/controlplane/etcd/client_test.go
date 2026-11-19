package etcd

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sitehub/common/registry"
)

type memoryWatch struct {
	key     string
	prefix  bool
	updates chan clientv3.WatchResponse
}

type memoryClient struct {
	mu        sync.Mutex
	revision  int64
	values    map[string]*mvccpb.KeyValue
	history   []*clientv3.Event
	watches   map[*memoryWatch]struct{}
	leases    map[clientv3.LeaseID]chan *clientv3.LeaseKeepAliveResponse
	nextLease clientv3.LeaseID
	failReads bool
	closed    bool
	afterGet  func()
}

func newMemoryClient() *memoryClient {
	return &memoryClient{revision: 1, values: make(map[string]*mvccpb.KeyValue), watches: make(map[*memoryWatch]struct{}), leases: make(map[clientv3.LeaseID]chan *clientv3.LeaseKeepAliveResponse)}
}

func (c *memoryClient) Get(ctx context.Context, key string, options ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	if c.failReads {
		c.mu.Unlock()
		return nil, errors.New("transport failure with credentials must never escape")
	}
	op := clientv3.OpGet(key, options...)
	response := &clientv3.GetResponse{Header: &etcdserverpb.ResponseHeader{Revision: c.revision}}
	for name, pair := range c.values {
		if matches(name, key, len(op.RangeBytes()) > 0) {
			copy := *pair
			response.Kvs = append(response.Kvs, &copy)
		}
	}
	hook := c.afterGet
	c.afterGet = nil
	c.mu.Unlock()
	if hook != nil {
		hook()
	}
	return response, nil
}

func matches(name, key string, prefix bool) bool {
	return name == key || (prefix && len(name) >= len(key) && name[:len(key)] == key)
}

func (c *memoryClient) Put(ctx context.Context, key, value string, options ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.revision++
	// This transport only receives WithLease on registration writes. The
	// production wire-level lease binding is covered by the real-server test.
	var lease int64
	if len(options) > 0 {
		lease = int64(c.nextLease)
	}
	pair := &mvccpb.KeyValue{Key: []byte(key), Value: []byte(value), ModRevision: c.revision, Lease: lease}
	c.values[key] = pair
	c.publish(&clientv3.Event{Type: mvccpb.PUT, Kv: pair})
	return &clientv3.PutResponse{Header: &etcdserverpb.ResponseHeader{Revision: c.revision}}, nil
}

func (c *memoryClient) publish(event *clientv3.Event) {
	c.history = append(c.history, event)
	for watch := range c.watches {
		if matches(string(event.Kv.Key), watch.key, watch.prefix) {
			watch.updates <- clientv3.WatchResponse{Header: etcdserverpb.ResponseHeader{Revision: c.revision}, Events: []*clientv3.Event{event}}
		}
	}
}

func (c *memoryClient) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.revision++
	delete(c.values, key)
	c.publish(&clientv3.Event{Type: mvccpb.DELETE, Kv: &mvccpb.KeyValue{Key: []byte(key), ModRevision: c.revision}})
}

func (c *memoryClient) Watch(ctx context.Context, key string, options ...clientv3.OpOption) clientv3.WatchChan {
	c.mu.Lock()
	op := clientv3.OpGet(key, options...)
	watch := &memoryWatch{key: key, prefix: len(op.RangeBytes()) > 0, updates: make(chan clientv3.WatchResponse, 256)}
	c.watches[watch] = struct{}{}
	for _, event := range c.history {
		if event.Kv.ModRevision >= op.Rev() && matches(string(event.Kv.Key), key, watch.prefix) {
			watch.updates <- clientv3.WatchResponse{Header: etcdserverpb.ResponseHeader{Revision: c.revision}, Events: []*clientv3.Event{event}}
		}
	}
	c.mu.Unlock()
	go func() { <-ctx.Done(); c.mu.Lock(); delete(c.watches, watch); close(watch.updates); c.mu.Unlock() }()
	return watch.updates
}

func (c *memoryClient) Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nextLease++
	c.leases[c.nextLease] = make(chan *clientv3.LeaseKeepAliveResponse, 2)
	return &clientv3.LeaseGrantResponse{ID: c.nextLease, TTL: ttl}, nil
}

func (c *memoryClient) Revoke(_ context.Context, lease clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ch := c.leases[lease]; ch != nil {
		close(ch)
		delete(c.leases, lease)
	}
	for key, value := range c.values {
		if value.Lease == int64(lease) {
			c.revision++
			delete(c.values, key)
			c.publish(&clientv3.Event{Type: mvccpb.DELETE, Kv: &mvccpb.KeyValue{Key: []byte(key), ModRevision: c.revision}})
		}
	}
	return &clientv3.LeaseRevokeResponse{}, nil
}

func (c *memoryClient) KeepAlive(ctx context.Context, lease clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := c.leases[lease]
	ch <- &clientv3.LeaseKeepAliveResponse{ID: lease, TTL: 15}
	return ch, nil
}

func (c *memoryClient) Close() error { c.mu.Lock(); c.closed = true; c.mu.Unlock(); return nil }

func testOptions() Options {
	return Options{Prefix: "/test/", RequestTimeout: 100 * time.Millisecond, LeaseTTL: 15}
}

func receiveValue(t *testing.T, changes <-chan string, want string) {
	t.Helper()
	select {
	case got := <-changes:
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for %q", want)
	}
}

func TestConfigWatchClosesReadWatchRaceAndDeletion(t *testing.T) {
	transport := newMemoryClient()
	_, _ = transport.Put(context.Background(), "/test/api.yaml", "Runtime: old")
	transport.afterGet = func() {
		_, _ = transport.Put(context.Background(), "/test/api.yaml", "Runtime: intermediate")
		_, _ = transport.Put(context.Background(), "/test/api.yaml", "Runtime: newest")
	}
	center := &ConfigCenter{base: newBase(transport, testOptions())}
	defer center.Close()
	changes := make(chan string, 10)
	stop, err := center.Watch(context.Background(), "api.yaml", func(data []byte, err error) {
		if err != nil {
			changes <- "error"
		} else {
			changes <- string(data)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	receiveValue(t, changes, "Runtime: old")
	receiveValue(t, changes, "Runtime: intermediate")
	receiveValue(t, changes, "Runtime: newest")
	transport.delete("/test/api.yaml")
	receiveValue(t, changes, "")
	if _, err := center.Get(context.Background(), "api.yaml"); err == nil {
		t.Fatal("missing config was accepted")
	}
}

func TestConfigWatchRecoversCompactedStream(t *testing.T) {
	transport := newMemoryClient()
	_, _ = transport.Put(context.Background(), "/test/api.yaml", "first")
	center := &ConfigCenter{base: newBase(transport, testOptions())}
	changes := make(chan string, 10)
	stop, err := center.Watch(context.Background(), "api.yaml", func(data []byte, err error) {
		if err != nil {
			changes <- "error"
		} else {
			changes <- string(data)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	receiveValue(t, changes, "first")
	// Wait for the watch to be installed, then simulate compacted history.
	deadline := time.Now().Add(time.Second)
	for {
		transport.mu.Lock()
		ready := len(transport.watches) > 0
		if ready {
			for watch := range transport.watches {
				watch.updates <- clientv3.WatchResponse{Canceled: true, CompactRevision: transport.revision}
			}
		}
		transport.mu.Unlock()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("watch not installed")
		}
		time.Sleep(time.Millisecond)
	}
	receiveValue(t, changes, "error")
	_, _ = transport.Put(context.Background(), "/test/api.yaml", "after compaction")
	receiveValue(t, changes, "after compaction")
	stop()
	stop()
	if err := center.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := center.Get(context.Background(), "api.yaml"); err == nil {
		t.Fatal("read after close")
	}
	transport.mu.Lock()
	closed := transport.closed
	transport.mu.Unlock()
	if !closed {
		t.Fatal("transport leaked")
	}
}

func TestWatchCancelInsideCallback(t *testing.T) {
	transport := newMemoryClient()
	center := &ConfigCenter{base: newBase(transport, testOptions())}
	ready, canceled := make(chan struct{}), make(chan struct{})
	var stop func()
	var err error
	stop, err = center.Watch(context.Background(), "api.yaml", func(_ []byte, _ error) { <-ready; stop(); close(canceled) })
	if err != nil {
		t.Fatal(err)
	}
	close(ready)
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Fatal("callback cancellation deadlocked")
	}
	if err := center.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRegistrationReestablishesLostLeaseAndCannotReviveAfterClose(t *testing.T) {
	transport := newMemoryClient()
	backend := &Registry{base: newBase(transport, testOptions())}
	instance := registry.Instance{Service: "test-rpc", IP: "127.0.0.1", Port: 9801, Healthy: true, Enabled: true, Weight: 100}
	registration, err := backend.Register(context.Background(), instance)
	if err != nil {
		t.Fatal(err)
	}
	transport.mu.Lock()
	initialLease := transport.nextLease
	transport.mu.Unlock()
	_, _ = transport.Revoke(context.Background(), initialLease)
	deadline := time.Now().Add(3 * time.Second)
	for {
		transport.mu.Lock()
		recovered := transport.nextLease > initialLease && len(transport.values) == 1
		transport.mu.Unlock()
		if recovered {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("lease was not reestablished")
		}
		time.Sleep(time.Millisecond)
	}
	if err := registration.Close(); err != nil {
		t.Fatal(err)
	}
	if err := registration.Close(); err != nil {
		t.Fatal(err)
	}
	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}
	transport.mu.Lock()
	defer transport.mu.Unlock()
	if len(transport.values) != 0 || len(transport.leases) != 0 {
		t.Fatal("registration or lease survived close")
	}
}
