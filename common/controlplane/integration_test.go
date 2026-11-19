package controlplane_test

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"sitehub/common/controlplane"
	"sitehub/common/registry"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestRegistryProxyResolvesRealRPCAndWithdraws(t *testing.T) {
	endpoint := os.Getenv("ETCD_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("set ETCD_TEST_ENDPOINT to run the real-server RPC proxy check")
	}
	c := controlplane.Config{
		Provider: "etcd", Endpoints: strings.Split(endpoint, ","), Prefix: "/sitehub-test/" + uuid.NewString(),
		Namespace: "integration", Group: "services", DialTimeout: time.Second,
		RequestTimeout: time.Second, LeaseTTL: 5,
	}
	provider, err := controlplane.NewRegistry(c)
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	go server.Serve(listener)
	defer server.Stop()
	port := uint64(listener.Addr().(*net.TCPAddr).Port)
	registration, err := provider.Register(context.Background(), registry.Instance{
		Service: "health-rpc", IP: "127.0.0.1", Port: port, Weight: 100, Enabled: true, Healthy: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer registration.Close()
	builder := &registry.Builder{Naming: provider, MaxStale: time.Second}
	defer builder.Close()
	client, err := zrpc.NewClient(zrpc.RpcClientConf{Target: "discovery:///health-rpc", NonBlock: true, Timeout: 500},
		zrpc.WithDialOption(grpc.WithResolvers(builder)))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Conn().Close()
	rpc := grpc_health_v1.NewHealthClient(client.Conn())
	check := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()
		_, err := rpc.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		return err
	}
	wait := func(description string, test func() bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if test() {
				return
			}
			time.Sleep(20 * time.Millisecond)
		}
		t.Fatal(description)
	}
	wait("RPC did not resolve the registered address", func() bool { return check() == nil })
	if err = registration.Close(); err != nil {
		t.Fatal(err)
	}
	// The server remains alive: failure must come from discovery withdrawal.
	wait("RPC kept using an address after authoritative withdrawal", func() bool { return check() != nil })
	registration, err = provider.Register(context.Background(), registry.Instance{
		Service: "health-rpc", IP: "127.0.0.1", Port: port, Weight: 100, Enabled: true, Healthy: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer registration.Close()
	wait("RPC did not recover after re-registration", func() bool { return check() == nil })
}
