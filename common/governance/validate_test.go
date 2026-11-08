package governance

import (
	"github.com/zeromicro/go-zero/zrpc"
	"testing"
	"time"
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
		c := config{Mode: "test", RPC: zrpc.RpcClientConf{Target: "nacos:///blog-rpc"}}
		c.Shutdown.WrapUpTime = 10 * time.Second
		c.Shutdown.WaitTime = 20 * time.Second
		c.Registry.Enabled = true
		c.Nacos.TimeoutMs = 5000
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
	c.Nacos.TimeoutMs = 20000
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
