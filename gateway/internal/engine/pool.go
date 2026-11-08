package engine

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"sitehub/gateway/internal/discovery"
	"sitehub/gateway/manifest"
)

type endpoint struct {
	target           *url.URL
	weight, failures int
	healthy          bool
}
type pool struct {
	config            manifest.Upstream
	transport         *http.Transport
	mutationTransport *http.Transport
	mu                sync.Mutex
	nodes             map[string]*endpoint
	cursor            atomic.Uint64
	cancel            context.CancelFunc
	unwatch           func()
	refs              int // guarded by Engine.mu
}

func newPool(config manifest.Upstream, provider *discovery.Manager) (*pool, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &pool{config: config, nodes: make(map[string]*endpoint), cancel: cancel}
	p.transport = &http.Transport{
		Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			conn, err := (&net.Dialer{Timeout: time.Duration(config.ConnectTimeoutMS) * time.Millisecond, KeepAlive: 30 * time.Second}).DialContext(ctx, network, address)
			if err != nil {
				return nil, err
			}
			return &idleConn{Conn: conn, timeout: time.Duration(config.IOTimeoutMS) * time.Millisecond}, nil
		},
		ForceAttemptHTTP2: false, TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{}, MaxIdleConns: 256, MaxIdleConnsPerHost: 64,
		IdleConnTimeout: 90 * time.Second, TLSHandshakeTimeout: time.Duration(config.ConnectTimeoutMS) * time.Millisecond,
		ResponseHeaderTimeout: time.Duration(config.ResponseHeaderTimeoutMS) * time.Millisecond,
		ExpectContinueTimeout: time.Second, DisableCompression: true,
	}
	p.mutationTransport = p.transport.Clone()
	p.mutationTransport.DisableKeepAlives = true
	if config.Service != "" {
		if provider == nil {
			p.close()
			return nil, fmt.Errorf("Nacos discovery is required for %s", config.Service)
		}
		stop, err := provider.Watch(config.Service, p.update)
		if err != nil {
			p.close()
			return nil, err
		}
		p.unwatch = stop
	} else {
		nodes := make([]discovery.Node, 0, len(config.Nodes))
		for _, n := range config.Nodes {
			nodes = append(nodes, discovery.Node{URL: n.URL, Weight: n.Weight})
		}
		p.update(nodes)
	}
	go p.monitor(ctx)
	return p, nil
}
func (p *pool) update(nodes []discovery.Node) {
	p.mu.Lock()
	defer p.mu.Unlock()
	next := make(map[string]*endpoint, len(nodes))
	for _, n := range nodes {
		target, err := url.Parse(n.URL)
		if err != nil || target.Host == "" || n.Weight <= 0 {
			continue
		}
		if old, ok := p.nodes[n.URL]; ok {
			old.weight = n.Weight
			next[n.URL] = old
		} else {
			next[n.URL] = &endpoint{target: target, weight: n.Weight}
		}
	}
	p.nodes = next // An authoritative empty discovery list removes every node.
}
func (p *pool) selectTarget() (*url.URL, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	keys := make([]string, 0, len(p.nodes))
	total := uint64(0)
	for key, n := range p.nodes {
		if n.healthy {
			keys = append(keys, key)
			total += uint64(n.weight)
		}
	}
	if total == 0 {
		return nil, false
	}
	sort.Strings(keys)
	position := (p.cursor.Add(1) - 1) % total
	for _, key := range keys {
		n := p.nodes[key]
		if position < uint64(n.weight) {
			copy := *n.target
			return &copy, true
		}
		position -= uint64(n.weight)
	}
	return nil, false
}
func (p *pool) monitor(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(p.config.HealthIntervalMS) * time.Millisecond)
	defer ticker.Stop()
	for {
		p.check(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (p *pool) check(ctx context.Context) {
	p.mu.Lock()
	targets := make(map[string]*endpoint, len(p.nodes))
	for k, n := range p.nodes {
		targets[k] = n
	}
	p.mu.Unlock()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 8)
	for key, node := range targets {
		select {
		case semaphore <- struct{}{}:
		case <-ctx.Done():
			wg.Wait()
			return
		}
		wg.Add(1)
		go func(key string, node *endpoint) {
			defer wg.Done()
			defer func() { <-semaphore }()
			target := *node.target
			target.Path = p.config.HealthPath
			target.RawPath = ""
			requestCtx, cancel := context.WithTimeout(ctx, time.Duration(p.config.HealthTimeoutMS)*time.Millisecond)
			defer cancel()
			req, _ := http.NewRequestWithContext(requestCtx, http.MethodGet, target.String(), nil)
			resp, err := p.transport.RoundTrip(req)
			ok := err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300
			if resp != nil {
				_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
				_ = resp.Body.Close()
			}
			p.mu.Lock()
			defer p.mu.Unlock()
			// A stale probe may not reinsert an endpoint removed by discovery.
			if current := p.nodes[key]; current == node {
				if ok {
					current.failures = 0
					current.healthy = true
				} else {
					current.failures++
					if current.failures >= p.config.HealthFailures {
						current.healthy = false
					}
				}
			}
		}(key, node)
	}
	wg.Wait()
}
func (p *pool) close() {
	p.cancel()
	if p.unwatch != nil {
		p.unwatch()
	}
	p.transport.CloseIdleConnections()
	if p.mutationTransport != nil {
		p.mutationTransport.CloseIdleConnections()
	}
}
