package etcd

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"sitehub/common/registry"
)

// Run with ETCD_TEST_ENDPOINT pointed at a disposable local etcd. Every test
// owns a random prefix; it never accesses an application's keys.
func TestEtcdIntegrationConfigAndRegistration(t *testing.T) {
	endpoint := os.Getenv("ETCD_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("set ETCD_TEST_ENDPOINT to run the real-server control-plane check")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	raw, err := clientv3.New(clientv3.Config{Endpoints: strings.Split(endpoint, ","), DialTimeout: time.Second, Logger: zap.NewNop()})
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	prefix := "/sitehub-test/" + uuid.NewString() + "/"
	defer func() {
		clean, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		_, _ = raw.Delete(clean, prefix, clientv3.WithPrefix())
	}()
	opts := Options{Endpoints: strings.Split(endpoint, ","), Prefix: prefix + "config/", DialTimeout: time.Second, RequestTimeout: time.Second, LeaseTTL: 5}
	center, err := NewConfigCenter(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer center.Close()
	if _, err := center.Get(ctx, "api.yaml"); err == nil {
		t.Fatal("missing config accepted")
	}
	_, err = raw.Put(ctx, opts.Prefix+"api.yaml", "Runtime: initial")
	if err != nil {
		t.Fatal(err)
	}
	changes := make(chan string, 20)
	stop, err := center.Watch(ctx, "api.yaml", func(data []byte, err error) {
		if err == nil {
			changes <- string(data)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	receiveValue(t, changes, "Runtime: initial")
	if _, err = raw.Put(ctx, opts.Prefix+"api.yaml", "Runtime: changed"); err != nil {
		t.Fatal(err)
	}
	receiveValue(t, changes, "Runtime: changed")
	if _, err = raw.Delete(ctx, opts.Prefix+"api.yaml"); err != nil {
		t.Fatal(err)
	}
	receiveValue(t, changes, "")
	opts.Prefix = prefix + "registry/"
	provider, err := NewRegistry(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	counts := make(chan int, 30)
	stopRegistry, err := provider.Watch(ctx, "test-rpc", func(instances []registry.Instance, err error) {
		if err == nil {
			counts <- len(instances)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stopRegistry()
	waitCount := func(want int) {
		t.Helper()
		select {
		case got := <-counts:
			if got != want {
				t.Fatalf("instances=%d,want %d", got, want)
			}
		case <-ctx.Done():
			t.Fatal("registry update timed out")
		}
	}
	waitCount(0)
	instance := registry.Instance{Service: "test-rpc", IP: "127.0.0.1", Port: 9801, Weight: 100, Healthy: true, Enabled: true}
	first, err := provider.Register(ctx, instance)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	waitCount(1)
	instance.Port = 9802
	second, err := provider.Register(ctx, instance)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	waitCount(2)
	if err = second.Close(); err != nil {
		t.Fatal(err)
	}
	waitCount(1)
	// Revoke the live lease externally; the provider must acquire a new lease
	// and republish the same registration, rather than silently disappearing.
	response, err := raw.Get(ctx, opts.Prefix, clientv3.WithPrefix())
	if err != nil || len(response.Kvs) != 1 {
		t.Fatalf("registered keys: %v %v", response, err)
	}
	if response.Kvs[0].Lease == 0 {
		t.Fatal("registration has no lease")
	}
	if _, err = raw.Revoke(ctx, clientv3.LeaseID(response.Kvs[0].Lease)); err != nil {
		t.Fatal(err)
	}
	waitCount(0)
	waitCount(1)
	if err = first.Close(); err != nil {
		t.Fatal(err)
	}
	waitCount(0)
}
