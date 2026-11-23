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
	"net/netip"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"sitehub/gateway/internal/clientaddr"
	"sitehub/gateway/internal/discovery"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

type Engine struct {
	trustedProxies  clientaddr.Trusted
	metrics         *metrics
	mu              sync.Mutex
	current         atomic.Pointer[Snapshot]
	registry        *plugin.Registry
	provider        *discovery.Manager
	pools           map[string]*pool
	states          map[string]plugin.Middleware
	maxBody         int64
	closed          bool
	streamSockets   map[string]*streamSocket
	streamContext   context.Context
	stopStreams     context.CancelFunc
	streamWorkers   sync.WaitGroup
	streamDraining  bool
	streamDrainDone chan struct{}
	streamSlots     chan struct{}
}
type route struct {
	config  manifest.Route
	host    manifest.HostMatcher
	handler http.Handler
}
type Snapshot struct {
	Document      manifest.Document
	Revision      string
	routes        []route
	pathWarnings  []string
	pools         map[string]*pool
	states        map[string]plugin.Middleware
	refs          atomic.Int64
	owner         *Engine
	streamTargets map[*streamSocket]string
	heldSockets   []*streamSocket // owned by the prepared configuration, then by the active revision
}

func New(provider *discovery.Manager, registry *plugin.Registry, maxBody int64, trusted ...netip.Prefix) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	return &Engine{metrics: newMetrics(), provider: provider, registry: registry, pools: map[string]*pool{}, states: map[string]plugin.Middleware{}, maxBody: maxBody, trustedProxies: trusted,
		streamSockets: make(map[string]*streamSocket), streamContext: ctx, stopStreams: cancel, streamSlots: make(chan struct{}, 1024)}
}
func (e *Engine) Prepare(d manifest.Document) (*Snapshot, error) {
	if err := manifest.Validate(d, e.registry); err != nil {
		return nil, err
	}
	// Own all configuration memory, including RawMessage byte slices.
	d, _ = manifest.Decode(manifest.Bytes(d))
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.streamDraining {
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
		host, _ := manifest.ParseHost(r.Host) // checked by manifest.Validate
		s.routes = append(s.routes, route{config: r, host: host, handler: middleware(terminal)})
	}
	sort.SliceStable(s.routes, func(i, j int) bool {
		if s.routes[i].config.Priority != s.routes[j].config.Priority {
			return s.routes[i].config.Priority > s.routes[j].config.Priority
		}
		if a, b := s.routes[i].host.Specificity(), s.routes[j].host.Specificity(); a != b {
			return a > b
		}
		return s.routes[i].config.ID < s.routes[j].config.ID
	})
	s.pathWarnings = manifest.PathWarnings(d.Routes)
	if err := e.reserveStreams(s); err != nil {
		return fail(err)
	}
	return s, nil
}
func (e *Engine) Commit(s *Snapshot) bool {
	e.mu.Lock()
	if e.closed || e.streamDraining {
		e.mu.Unlock()
		s.Release()
		return false
	}
	e.metrics.reloads.Inc()
	e.states = s.states
	previous := e.current.Swap(s)
	for _, socket := range s.heldSockets {
		if !socket.started {
			socket.started = true
			e.streamWorkers.Add(1)
			go socket.serve()
		}
	}
	if previous != nil {
		e.releaseStreams(previous)
	}
	e.mu.Unlock()
	if previous != nil {
		previous.Release()
	}
	return true
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
	e.releaseStreams(s)
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
	e.stopStreams()
	if previous != nil {
		e.releaseStreams(previous)
	}
	// A publisher can still hold an uncommitted reservation during shutdown.
	for _, socket := range e.streamSockets {
		socket.close()
	}
	e.mu.Unlock()
	if previous != nil {
		previous.Release()
	}
	e.streamWorkers.Wait()
}
func (e *Engine) Revision() string {
	if s := e.current.Load(); s != nil {
		return s.Revision
	}
	return ""
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	forwarded := e.trustedProxies.Resolve(r)
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
	selected, _, _, err := s.match(r.Host, r.Method, forwarded.Scheme, r.URL)
	if err != nil {
		writeError(w, 400, err.Error())
		return
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
	r = r.Clone(forwarded.Context(r.Context()))
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
			info := clientaddr.From(req.In)
			req.Out.Header.Set("X-Forwarded-For", info.Client)
			req.Out.Header.Set("X-Real-IP", info.Client)
			req.Out.Header.Set("X-Forwarded-Proto", info.Scheme)
			req.Out.Header.Set("X-Forwarded-Host", info.Host)
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
