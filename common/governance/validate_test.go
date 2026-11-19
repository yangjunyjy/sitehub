package governance

import (
	"strings"
	"testing"
	"time"

	"sitehub/common/controlplane"

	"github.com/zeromicro/go-zero/zrpc"
)

func TestRegistryRejectsCompetingDiscoveryAndShortShutdown(t *testing.T) {
	type config struct {
		Config
		Mode string
		// go-zero's Windows ShutdownConf is intentionally empty. Exercise the
		// production layout here; the Linux fixture validates real signal hooks.
		Shutdown struct{ WrapUpTime, WaitTime time.Duration }
		RPC      zrpc.RpcClientConf
	}
	fresh := func() config {
		c := config{Mode: "test", RPC: zrpc.RpcClientConf{Target: "discovery:///rbac-rpc"}}
		c.Shutdown.WrapUpTime = 20 * time.Second
		c.Shutdown.WaitTime = 75 * time.Second
		c.Registry.Enabled = true
		c.ControlPlane = controlplane.Config{Provider: "etcd", Endpoints: []string{"127.0.0.1:2379"}, Prefix: "/sitehub", Namespace: "test", Group: "SITEHUB", DialTimeout: time.Second, RequestTimeout: 5 * time.Second, LeaseTTL: 15}
		return c
	}
	c := fresh()
	if err := Validate(c, c.Config); err != nil {
		t.Fatal(err)
	}
	c.RPC.Endpoints = []string{"localhost:9802"}
	if Validate(c, c.Config) == nil {
		t.Fatal("Endpoints silently override discovery")
	}
	c = fresh()
	c.Shutdown.WaitTime = 5 * time.Second
	if Validate(c, c.Config) == nil {
		t.Fatal("process can exit before deregistration")
	}
	c = fresh()
	c.ControlPlane.RequestTimeout = 20 * time.Second
	if Validate(c, c.Config) == nil {
		t.Fatal("registry timeout exceeds withdrawal window")
	}
	c = fresh()
	c.Mode = "pro"
	c.Registry.AllowLoopback = true
	if Validate(c, c.Config) == nil {
		t.Fatal("production loopback accepted")
	}
}

func TestShutdownWindowIncludesProviderTimeoutAndLeaseExpiry(t *testing.T) {
	for _, test := range []struct {
		name     string
		timeout  time.Duration
		leaseTTL int64
		wrap     time.Duration
		minimum  string
	}{
		{name: "minimum propagation window", timeout: time.Second, leaseTTL: 5, wrap: 9 * time.Second, minimum: "10s"},
		{name: "deployed lease timeout", timeout: 5 * time.Second, leaseTTL: 15, wrap: 10 * time.Second, minimum: "20s"},
		{name: "longer provider timeout", timeout: 30 * time.Second, leaseTTL: 15, wrap: 20 * time.Second, minimum: "35s"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var c struct {
				Config
				Shutdown struct{ WrapUpTime, WaitTime time.Duration }
			}
			c.Registry.Enabled = true
			c.ControlPlane = controlplane.Config{Provider: "etcd", Endpoints: []string{"127.0.0.1:2379"}, Prefix: "/sitehub", Namespace: "test", Group: "SITEHUB", DialTimeout: time.Second, RequestTimeout: test.timeout, LeaseTTL: test.leaseTTL}
			c.Shutdown.WrapUpTime = test.wrap
			c.Shutdown.WaitTime = 75 * time.Second
			err := Validate(c, c.Config)
			if err == nil || !strings.Contains(err.Error(), "at least "+test.minimum) || !strings.Contains(err.Error(), "configured "+test.wrap.String()) {
				t.Fatalf("expected required and configured shutdown windows, got %v", err)
			}
			minimum, err := time.ParseDuration(test.minimum)
			if err != nil {
				t.Fatal(err)
			}
			c.Shutdown.WrapUpTime = minimum
			if err := Validate(c, c.Config); err != nil {
				t.Fatalf("exact required shutdown window was rejected: %v", err)
			}
		})
	}
}
