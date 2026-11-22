package app

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"sitehub/common/controlplane"
	"sitehub/gateway/console"
	"sitehub/gateway/internal/config"
	"sitehub/gateway/internal/control"
	"sitehub/gateway/internal/discovery"
	"sitehub/gateway/internal/engine"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
	"sitehub/gateway/publisher"

	"github.com/zeromicro/go-zero/core/logx"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Run(ctx context.Context, c config.Config) error {
	return RunWithRegistry(ctx, c, plugin.DefaultRegistry())
}

// RunWithRegistry lets an embedding gateway register compiled plugins once and
// use the same schemas for file loading, validation and request execution.
func RunWithRegistry(ctx context.Context, c config.Config, plugins *plugin.Registry) error {
	if plugins == nil {
		return fmt.Errorf("plugin registry is required")
	}
	if err := c.Validate(); err != nil {
		return err
	}
	allowed, err := control.AllowedCIDRs(c.Admin.AllowedCIDRs)
	if err != nil {
		return err
	}
	load := func() (manifest.Document, error) { return ReadRoutes(c, plugins) }
	seed, err := Check(c, plugins)
	if err != nil {
		return err
	}
	if err = logx.SetUp(c.Log); err != nil {
		return err
	}
	defer logx.Close()
	var store control.Store = &control.MemoryStore{}
	var watch func(context.Context, func(control.Record), func(error))
	if c.SourceType() == "etcd" {
		storeClient, err := clientv3.New(clientv3.Config{
			Endpoints:   c.Etcd.Endpoints,
			Username:    c.Etcd.Username,
			Password:    c.Etcd.Password,
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			return fmt.Errorf("connect to gateway configuration store failed")
		}
		defer storeClient.Close()
		shared := &control.EtcdStore{Client: storeClient, Key: c.Etcd.Key}
		store, watch = shared, shared.Watch
	}
	var provider *discovery.Manager
	if c.Discovery.Enabled {
		registry, err := controlplane.NewRegistry(c.ControlPlane)
		if err != nil {
			return err
		}
		defer registry.Close()
		provider = discovery.New(registry, c.Discovery.MaxStale)
		defer provider.Close()
	}
	proxy := engine.New(provider, plugins, c.HTTP.MaxBodyBytes)
	defer proxy.Close()
	controller := &control.Controller{Store: store, Engine: proxy, Key: c.Admin.Key, Allowed: allowed}
	if c.SourceType() == "file" {
		controller.Load = load
	}
	// Reserve the process endpoints before a stream manifest can claim any of
	// them or become the first persisted configuration.
	publicListener, err := net.Listen("tcp", c.Listen)
	if err != nil {
		return err
	}
	defer publicListener.Close()
	adminListener, err := net.Listen("tcp", c.Admin.Listen)
	if err != nil {
		return err
	}
	defer adminListener.Close()
	var consoleListener net.Listener
	if c.Console.Enabled {
		consoleListener, err = net.Listen("tcp", c.Console.Listen)
		if err != nil {
			return err
		}
		defer consoleListener.Close()
	}
	// Validate the seed before the first persistent write. Existing records are
	// authoritative and never overwritten by process startup or image upgrades.
	record, err := store.Read(ctx)
	if err != nil {
		return err
	}
	if record.Revision == 0 {
		prepared, err := proxy.Prepare(seed)
		if err != nil {
			return err
		}
		// Keep candidate sockets reserved through the first Apply. It shares
		// unchanged listeners, including after a concurrent seed wins the CAS.
		record, err = control.Seed(ctx, store, seed)
		if err == nil && string(record.Data) == string(manifest.Bytes(seed)) {
			err = controller.Apply(record)
		}
		prepared.Release()
		if err != nil {
			return err
		}
	}
	if err = controller.Apply(record); err != nil {
		return err
	}
	watchCtx, cancelWatch := context.WithCancel(context.Background())
	watchDone := make(chan struct{})
	// The watcher can prepare snapshots using both the engine and the registry
	// proxy. Join it before deferred dependency cleanup on every exit path.
	defer func() { cancelWatch(); <-watchDone }()
	go func() {
		defer close(watchDone)
		if watch == nil {
			return
		}
		watch(watchCtx, func(r control.Record) {
			if err := controller.Apply(r); err != nil {
				logx.Errorw("gateway configuration rejected", logx.Field("reason", err.Error()))
			}
		}, controller.Failure)
	}()
	connections := newConnections()
	public := &http.Server{Addr: c.Listen, Handler: console.WithoutSession(proxy), ReadHeaderTimeout: c.HTTP.ReadHeaderTimeout, IdleTimeout: c.HTTP.IdleTimeout, MaxHeaderBytes: 1 << 20, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	admin := &http.Server{Addr: c.Admin.Listen, Handler: controller.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	publicListener = &trackingListener{Listener: publicListener, owner: connections}
	failures := make(chan error, 3)
	servers := []*http.Server{public, admin}
	if c.Console.Enabled {
		host, port, _ := net.SplitHostPort(c.Admin.Listen)
		if host == "" || host == "0.0.0.0" {
			host = "127.0.0.1"
		} else if host == "::" {
			host = "::1"
		}
		client := &publisher.Client{URL: "http://" + net.JoinHostPort(host, port), Key: c.Admin.Key, HTTP: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
		handler, err := console.New(c.Console.Config, client, plugins, c.Console.StateDir)
		if err != nil {
			return err
		}
		web := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 40 * time.Second, IdleTimeout: 60 * time.Second}
		servers = append(servers, web)
		go func() { failures <- web.Serve(consoleListener) }()
	}
	go func() {
		if c.TLS.Certificate != "" {
			failures <- public.ServeTLS(publicListener, c.TLS.Certificate, c.TLS.Key)
		} else {
			failures <- public.Serve(publicListener)
		}
	}()
	go func() { failures <- admin.Serve(adminListener) }()
	logx.Infow("gateway started", logx.Field("listen", c.Listen), logx.Field("revision", record.Revision))
	var result error
	select {
	case <-ctx.Done():
	case result = <-failures:
		if errors.Is(result, http.ErrServerClosed) {
			result = nil
		}
	}
	controller.Stop()
	cancelWatch()
	drainCtx, cancel := context.WithTimeout(context.Background(), c.HTTP.ShutdownTimeout)
	defer cancel()
	var wg sync.WaitGroup
	for _, srv := range servers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := srv.Shutdown(drainCtx); err != nil {
				_ = srv.Close()
			}
		}()
	}
	wg.Wait()
	// Shutdown excludes hijacked WebSocket connections. Give them the same
	// drain deadline, then close tracked sockets to bound process termination.
	connections.drain(drainCtx)
	return result
}

type connections struct {
	mu      sync.Mutex
	active  map[*trackedConn]struct{}
	changed chan struct{}
	closed  bool
}

func newConnections() *connections {
	return &connections{active: make(map[*trackedConn]struct{}), changed: make(chan struct{})}
}

// Track below TLS and HTTP. net/http does not issue StateClosed after a
// connection is hijacked, whereas the underlying Conn.Close is always called
// by ReverseProxy when an upgraded session finishes.
type trackingListener struct {
	net.Listener
	owner *connections
}

func (l *trackingListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	l.owner.mu.Lock()
	if l.owner.closed {
		l.owner.mu.Unlock()
		_ = conn.Close()
		return nil, net.ErrClosed
	}
	tracked := &trackedConn{Conn: conn, owner: l.owner}
	l.owner.active[tracked] = struct{}{}
	l.owner.signalLocked()
	l.owner.mu.Unlock()
	return tracked, nil
}

type trackedConn struct {
	net.Conn
	owner *connections
	once  sync.Once
	err   error
}

func (c *trackedConn) Close() error {
	c.once.Do(func() {
		c.err = c.Conn.Close()
		c.owner.mu.Lock()
		delete(c.owner.active, c)
		c.owner.signalLocked()
		c.owner.mu.Unlock()
	})
	return c.err
}

func (c *connections) signalLocked() {
	close(c.changed)
	c.changed = make(chan struct{})
}

// Called after HTTP Shutdown has drained ordinary requests. Sessions that
// finish naturally wake this wait immediately; remaining upgraded sockets are
// closed at the same original shutdown deadline, not a second timeout.
func (c *connections) drain(ctx context.Context) {
	for {
		c.mu.Lock()
		if len(c.active) == 0 {
			c.closed = true
			c.mu.Unlock()
			return
		}
		changed := c.changed
		c.mu.Unlock()
		select {
		case <-changed:
		case <-ctx.Done():
			c.closeAll()
			return
		}
	}
}

func (c *connections) closeAll() {
	c.mu.Lock()
	c.closed = true
	pending := make([]*trackedConn, 0, len(c.active))
	for conn := range c.active {
		pending = append(pending, conn)
	}
	c.mu.Unlock()
	for _, conn := range pending {
		_ = conn.Close()
	}
}
