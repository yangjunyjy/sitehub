package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sitehub/common/controlplane"
	"sitehub/common/registry"
	"sitehub/common/runtimeconfig"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"
	"google.golang.org/grpc/resolver"
)

type contextKey struct{}
type Runtime struct {
	Name         string
	Config       Config
	Traffic      *runtimeconfig.Manager
	naming       *registry.Proxy
	resolver     *registry.Builder
	probe        func(context.Context) error
	ready        atomic.Bool
	mu           sync.Mutex
	closing      bool
	registration registry.Registration
	management   *http.Server
}

func New(name string, c Config, traffic *runtimeconfig.Manager) (*Runtime, error) {
	r := &Runtime{Name: name, Config: c, Traffic: traffic}
	if c.Registry.Enabled {
		if c.Registry.MaxStale < time.Second || c.Registry.StartupTimeout < time.Second {
			return nil, fmt.Errorf("registry timeouts must be at least 1s")
		}
		naming, err := controlplane.NewRegistry(c.ControlPlane)
		if err != nil {
			return nil, err
		}
		r.naming = naming
		r.resolver = &registry.Builder{Naming: naming, MaxStale: c.Registry.MaxStale}
		resolver.Register(r.resolver)
	}
	if c.Management.ListenOn != "" {
		listener, err := net.Listen("tcp", c.Management.ListenOn)
		if err != nil {
			r.Close()
			return nil, fmt.Errorf("listen on management address: %w", err)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", r.health)
		mux.HandleFunc("/readyz", r.readiness)
		mux.HandleFunc("/internal/governance", r.status)
		r.management = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
		go func() {
			if err := r.management.Serve(listener); err != nil && err != http.ErrServerClosed {
				logx.Error("management server stopped")
			}
		}()
	}
	// go-zero runs this phase before its HTTP/RPC shutdown listeners. Compose
	// sets WrapUpTime > the discovery refresh and WaitTime > the drain budget.
	proc.AddWrapUpListener(r.Withdraw)
	return r, nil
}
func WithContext(ctx context.Context, r *Runtime) context.Context {
	return context.WithValue(ctx, contextKey{}, r)
}
func FromContext(ctx context.Context) *Runtime {
	return ctx.Value(contextKey{}).(*Runtime)
}
func (r *Runtime) status(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(r.Traffic.Status())
}
func (r *Runtime) health(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
func (r *Runtime) readiness(w http.ResponseWriter, _ *http.Request) {
	if !r.ready.Load() || r.Traffic.Snapshot().Config.Maintenance || r.checkReady(context.Background()) != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Runtime) SetReadiness(probe func(context.Context) error) {
	r.mu.Lock()
	r.probe = probe
	r.mu.Unlock()
}
func (r *Runtime) checkReady(parent context.Context) error {
	r.mu.Lock()
	probe := r.probe
	r.mu.Unlock()
	if probe == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return probe(ctx)
}
func (r *Runtime) AttachHTTP(server *rest.Server) {
	server.Use(func(next http.HandlerFunc) http.HandlerFunc {
		limited := r.Traffic.HTTP(next)
		return func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/healthz", "/readyz", "/internal/governance":
				next(w, req)
			default:
				limited(w, req)
			}
		}
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/healthz",
		Handler: r.health,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/readyz",
		Handler: r.readiness,
	})
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/internal/governance",
		Handler: r.status,
	})
}

// Serve waits for the actual socket before registration. Dependency construction
// has already completed in ServiceContext. Registration errors fail startup.
func (r *Runtime) Serve(ctx context.Context, address string, start func()) error {
	finished := make(chan struct{})
	go func() { defer close(finished); start() }()
	if !r.Config.Registry.Enabled {
		r.ready.Store(true)
		<-finished
		return nil
	}
	ip, port, err := registry.ResolveAdvertise(r.Config.Registry.AdvertiseIP, address, r.Config.Registry.AllowLoopback)
	if err != nil {
		return err
	}
	host, _, _ := net.SplitHostPort(address)
	if host == "" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	probe := net.JoinHostPort(host, strconv.FormatUint(port, 10))
	deadline := time.NewTimer(r.Config.Registry.StartupTimeout)
	defer deadline.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-finished:
			return fmt.Errorf("service stopped before registration")
		case <-deadline.C:
			return fmt.Errorf("service did not listen before registration deadline")
		case <-tick.C:
			conn, err := net.DialTimeout("tcp", probe, 200*time.Millisecond)
			if err != nil {
				continue
			}
			_ = conn.Close()
			if err := r.checkReady(ctx); err != nil {
				continue
			}
			r.mu.Lock()
			if r.closing {
				r.mu.Unlock()
				return fmt.Errorf("service is shutting down")
			}
			protocol := "http"
			if strings.HasSuffix(r.Name, "-rpc") {
				protocol = "grpc"
			}
			registration, err := r.naming.Register(ctx, registry.Instance{Service: r.Name, IP: ip, Port: port, Weight: 100, Enabled: true, Healthy: true, Metadata: map[string]string{"protocol": protocol}})
			if err == nil {
				r.registration = registration
				r.ready.Store(true)
			}
			r.mu.Unlock()
			if err != nil {
				return err
			}
			<-finished
			return nil
		}
	}
}
func (r *Runtime) Ready() { r.ready.Store(true) } // Scheduler consumes RPCs, has no provider registration.
func (r *Runtime) Withdraw() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closing = true
	r.ready.Store(false)
	if r.registration != nil {
		if err := r.registration.Close(); err != nil {
			logx.Error(err)
		}
		r.registration = nil
	}
}
func (r *Runtime) Close() {
	r.Withdraw()
	if r.management != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = r.management.Shutdown(ctx)
	}
	if r.naming != nil {
		r.resolver.Close()
		_ = r.naming.Close()
	}
}
