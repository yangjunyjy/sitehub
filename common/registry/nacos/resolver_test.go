package nacos

import (
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc/resolver"
)

type fakeNaming struct {
	mu           sync.Mutex
	param        *vo.SubscribeParam
	healthy      bool
	instances    []model.Instance
	unsubscribed bool
}

func (f *fakeNaming) Subscribe(p *vo.SubscribeParam) error { f.param = p; return nil }
func (f *fakeNaming) Unsubscribe(p *vo.SubscribeParam) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unsubscribed = true
	return nil
}
func (f *fakeNaming) SelectInstances(vo.SelectInstancesParam) ([]model.Instance, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.instances, nil
}
func (f *fakeNaming) ServerHealthy() bool { f.mu.Lock(); defer f.mu.Unlock(); return f.healthy }

type captureConn struct {
	resolver.ClientConn
	states chan resolver.State
}

func (c *captureConn) UpdateState(s resolver.State) error { c.states <- s; return nil }
func (c *captureConn) ReportError(error)                  {}

func TestDiscoveryEmptyOutageAndUnsubscribe(t *testing.T) {
	instance := model.Instance{Ip: "127.0.0.1", Port: 9801, Healthy: true, Enable: true, Weight: 100}
	naming := &fakeNaming{healthy: true, instances: []model.Instance{instance}}
	cc := &captureConn{states: make(chan resolver.State, 20)}
	u, _ := url.Parse("nacos:///rbac-rpc")
	b := Builder{Naming: naming, Group: "SITEHUB", MaxStale: 40 * time.Millisecond}
	r, err := b.Build(resolver.Target{URL: *u}, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	next := func(want int) {
		t.Helper()
		select {
		case state := <-cc.states:
			if len(state.Addresses) != want {
				t.Fatalf("addresses=%v", state.Addresses)
			}
		case <-time.After(time.Second):
			t.Fatal("discovery update timed out")
		}
	}
	next(1)
	naming.param.SubscribeCallback(nil, nil)
	next(0)
	naming.param.SubscribeCallback([]model.Instance{instance}, nil)
	next(1)
	naming.param.SubscribeCallback(nil, errors.New("transport failure"))
	select {
	case <-cc.states:
		t.Fatal("transport error immediately replaced last good snapshot")
	default:
	}
	naming.mu.Lock()
	naming.healthy = false
	naming.mu.Unlock()
	next(0)
	naming.mu.Lock()
	naming.healthy = true
	naming.mu.Unlock()
	next(1)
	r.Close()
	if !naming.unsubscribed {
		t.Fatal("subscription leaked")
	}
	naming.param.SubscribeCallback([]model.Instance{instance}, nil)
	select {
	case <-cc.states:
		t.Fatal("update after close")
	default:
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
