// Package engine builds immutable routing snapshots. In-flight requests hold a
// reference to their snapshot; retirement releases resources only after drain.
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"sitehub/gateway/internal/discovery"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

type Engine struct {
	metrics  *metrics
	mu       sync.Mutex
	current  atomic.Pointer[Snapshot]
	registry *plugin.Registry
	provider *discovery.Manager
	pools    map[string]*pool
	states   map[string]plugin.Middleware
	maxBody  int64
	closed   bool
}
type route struct {
	config  manifest.Route
	handler http.Handler
}
type Snapshot struct {
	Document manifest.Document
	Revision string
	routes   []route
	pools    map[string]*pool
	states   map[string]plugin.Middleware
	refs     atomic.Int64
	owner    *Engine
}

func New(provider *discovery.Manager, registry *plugin.Registry, maxBody int64) *Engine {
	return &Engine{metrics: newMetrics(), provider: provider, registry: registry, pools: map[string]*pool{}, states: map[string]plugin.Middleware{}, maxBody: maxBody}
}
func (e *Engine) Prepare(d manifest.Document) (*Snapshot, error) {
	if err := manifest.Validate(d, e.registry); err != nil {
		return nil, err
	}
	// Own all configuration memory, including RawMessage byte slices.
	d, _ = manifest.Decode(manifest.Bytes(d))
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return nil, fmt.Errorf("gateway is stopping")
	}
	s := &Snapshot{Document: d, Revision: manifest.Revision(d), owner: e, pools: map[string]*pool{}, states: map[string]plugin.Middleware{}}
	s.refs.Store(1)
	fail := func(err error) (*Snapshot, error) { e.releasePools(s); return nil, err }
	for name, u := range d.Upstreams {
		data, _ := json.Marshal(u)
		key := string(data)
		p, ok := e.pools[key]
		if !ok {
			var err error
			p, err = newPool(u, e.provider)
			if err != nil {
				return fail(err)
			}
			e.pools[key] = p
		}
		p.refs++
		s.pools[name] = p
	}
	for _, r := range d.Routes {
		configs := manifest.EffectivePlugins(d, r)
		data, _ := json.Marshal(configs)
		key := r.ID + "|" + r.Host + "|" + r.Upstream + "|" + string(data)
		middleware, ok := e.states[key]
		if !ok {
			var err error
			middleware, err = e.registry.Build(configs)
			if err != nil {
				return fail(err)
			}
		}
		s.states[key] = middleware
		var terminal http.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeError(w, 503, "No upstream available") })
		if r.Upstream != "" {
			terminal = e.proxy(s.pools[r.Upstream])
		}
		s.routes = append(s.routes, route{config: r, handler: middleware(terminal)})
	}
	sort.SliceStable(s.routes, func(i, j int) bool {
		if s.routes[i].config.Priority != s.routes[j].config.Priority {
			return s.routes[i].config.Priority > s.routes[j].config.Priority
		}
		return s.routes[i].config.ID < s.routes[j].config.ID
	})
	return s, nil
}
func (e *Engine) Commit(s *Snapshot) {
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		s.Release()
		return
	}
	e.metrics.reloads.Inc()
	e.states = s.states
	previous := e.current.Swap(s)
	e.mu.Unlock()
	if previous != nil {
		previous.Release()
	}
}
func (s *Snapshot) acquire() bool {
	for {
		n := s.refs.Load()
		if n == 0 {
			return false
		}
		if s.refs.CompareAndSwap(n, n+1) {
			return true
		}
	}
}
func (s *Snapshot) Release() {
	if s.refs.Add(-1) == 0 {
		s.owner.mu.Lock()
		defer s.owner.mu.Unlock()
		s.owner.releasePools(s)
	}
}
func (e *Engine) releasePools(s *Snapshot) {
	for _, p := range s.pools {
		p.refs--
		if p.refs == 0 {
			data, _ := json.Marshal(p.config)
			delete(e.pools, string(data))
			p.close()
		}
	}
}
func (e *Engine) Close() {
	e.mu.Lock()
	e.closed = true
	previous := e.current.Swap(nil)
	e.mu.Unlock()
	if previous != nil {
		previous.Release()
	}
}
func (e *Engine) Revision() string {
	if s := e.current.Load(); s != nil {
		return s.Revision
	}
	return ""
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	observed := &accessWriter{ResponseWriter: w}
	w = observed
	routeID := "unmatched"
	e.metrics.active.Inc()
	defer func() { e.observe(observed, routeID, started) }()
	var s *Snapshot
	for {
		s = e.current.Load()
		if s == nil {
			writeError(w, 503, "Gateway not ready")
			return
		}
		if s.acquire() {
			break
		}
	}
	defer s.Release()
	// Reject ambiguous paths before matching, so proxy and application cannot
	// disagree about encoded separators, traversal, or internal control paths.
	p := r.URL.Path
	raw := strings.ToLower(r.URL.EscapedPath())
	if strings.ContainsAny(p, "\\\x00\r\n") || strings.Contains(raw, "%2f") || strings.Contains(raw, "%5c") || strings.Contains(raw, "%25") || strings.Contains(p, "//") || (path.Clean(p) != p && path.Clean(p)+"/" != p) {
		writeError(w, 400, "Invalid request path")
		return
	}
	host := strings.ToLower(r.Host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	var selected *route
	longest := -1
	for i := range s.routes {
		candidate := &s.routes[i]
		c := candidate.config
		if c.Host != host {
			continue
		}
		if len(c.Methods) > 0 {
			matched := false
			for _, m := range c.Methods {
				if m == r.Method {
					matched = true
				}
			}
			if !matched {
				continue
			}
		}
		for _, pattern := range c.Paths {
			match := p == pattern || (strings.HasSuffix(pattern, "/*") && strings.HasPrefix(p, strings.TrimSuffix(pattern, "*")))
			if match && (selected == nil || (c.Priority == selected.config.Priority && len(pattern) > longest)) {
				selected = candidate
				longest = len(pattern)
			}
		}
		if selected != nil && c.Priority < selected.config.Priority {
			break
		}
	}
	if selected == nil {
		writeError(w, 404, "Not found")
		return
	}
	routeID = selected.config.ID
	if r.ContentLength > e.maxBody {
		writeError(w, 413, "Request body too large")
		return
	}
	r = r.Clone(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, e.maxBody)
	for _, name := range []string{"X-Service-Key", "X-Actor-Id", "X-User-Id", "X-Role", "X-Roles", "Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto", "X-Real-Ip", "X-Original-Url", "X-Rewrite-Url"} {
		r.Header.Del(name)
	}
	timeout := 10 * time.Second
	if selected.config.Upstream != "" {
		timeout = time.Duration(s.pools[selected.config.Upstream].config.IOTimeoutMS) * time.Millisecond
	}
	control := http.NewResponseController(w)
	defer func() { _ = control.SetReadDeadline(time.Time{}); _ = control.SetWriteDeadline(time.Time{}) }()
	r.Body = &idleBody{ReadCloser: r.Body, control: control, timeout: timeout}
	selected.handler.ServeHTTP(&idleWriter{ResponseWriter: w, control: control, timeout: timeout}, r)
}

type targetKey struct{}

func (e *Engine) proxy(p *pool) http.Handler {
	proxy := &httputil.ReverseProxy{
		Transport: noRetryTransport{p.transport, p.mutationTransport}, FlushInterval: -1,
		ErrorLog: log.New(discardLog{}, "", 0),
		Rewrite: func(req *httputil.ProxyRequest) {
			target := req.In.Context().Value(targetKey{}).(targetValue)
			req.Out.URL.Scheme = target.scheme
			req.Out.URL.Host = target.host
			req.Out.Host = req.In.Host
			req.SetXForwarded()
			ip, _, _ := net.SplitHostPort(req.In.RemoteAddr)
			req.Out.Header.Set("X-Real-IP", ip)
			// Rewrite runs AFTER hop-by-hop header removal. Preserve validated
			// credentials even when a client names Authorization in Connection.
			for _, name := range []string{"Authorization", "X-Request-Id"} {
				if value := req.In.Header.Get(name); value != "" {
					req.Out.Header.Set(name, value)
				}
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if r.Context().Err() != nil {
				return
			}
			var bodyError *http.MaxBytesError
			if errors.As(err, &bodyError) {
				writeError(w, 413, "Request body too large")
				return
			}
			if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				writeError(w, 504, "Upstream timeout")
				return
			}
			writeError(w, 502, "Upstream unavailable")
		},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, ok := p.selectTarget()
		if !ok {
			writeError(w, 503, "No healthy upstream")
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), targetKey{}, targetValue{target.Scheme, target.Host}))
		proxy.ServeHTTP(w, r)
	})
}

type targetValue struct{ scheme, host string }

// Mutations use fresh HTTP/1 connections so net/http cannot replay them after
// a stale pooled-connection failure. Preserve application idempotency headers.
type noRetryTransport struct{ base, mutation *http.Transport }

func (t noRetryTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
		r.GetBody = nil
		return t.mutation.RoundTrip(r)
	}
	return t.base.RoundTrip(r)
}

type discardLog struct{}

func (discardLog) Write(p []byte) (int, error) { return len(p), nil }
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": status, "message": message})
}
