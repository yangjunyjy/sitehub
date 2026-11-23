// Package plugin supplies explicitly registered gateway middleware. Plugin code
// is compiled with the gateway; configuration can change without a restart.
package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
)

type Middleware func(http.Handler) http.Handler
type Factory func(json.RawMessage) (Middleware, error)

type registration struct {
	priority int
	factory  Factory
	terminal bool
}

type Registry struct {
	mu      sync.RWMutex
	plugins map[string]registration
}

func NewRegistry() *Registry { return &Registry{plugins: make(map[string]registration)} }

func (r *Registry) Register(name string, priority int, factory Factory) error {
	return r.register(name, priority, factory, false)
}

// RegisterTerminal registers a plugin that produces the response instead of
// forwarding to an upstream. Routing validation uses this metadata, so adding
// another response-producing plugin does not require changes to the router.
func (r *Registry) RegisterTerminal(name string, priority int, factory Factory) error {
	return r.register(name, priority, factory, true)
}

func (r *Registry) register(name string, priority int, factory Factory, terminal bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if name == "" || factory == nil {
		return fmt.Errorf("plugin name and factory are required")
	}
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	if r.plugins == nil {
		r.plugins = make(map[string]registration)
	}
	r.plugins[name] = registration{priority: priority, factory: factory, terminal: terminal}
	return nil
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	for _, item := range []struct {
		name     string
		priority int
		factory  Factory
	}{
		{"request-id", 1000, requestID},
		{"deny-paths", 950, denyPaths},
		{"browser-session", 900, browserSession},
		{"strip-credentials", 850, stripCredentials},
		{"rate-limit", 800, rateLimit},
		{"limit-count", 700, limitCount},
		{"proxy-rewrite", 600, proxyRewrite},
		{"circuit-breaker", 500, circuitBreaker},
	} {
		_ = r.Register(item.name, item.priority, item.factory)
	}
	_ = r.RegisterTerminal("static-response", 100, staticResponse)
	return r
}

func (r *Registry) Validate(configs map[string]json.RawMessage) error {
	_, err := r.Build(configs)
	return err
}

// ValidateOverrides checks a configuration layer before inheritance is
// resolved. An explicit null disables an inherited plugin, but its name must
// still be registered so spelling mistakes cannot silently remove protection.
func (r *Registry) ValidateOverrides(configs map[string]json.RawMessage) error {
	active := make(map[string]json.RawMessage, len(configs))
	r.mu.RLock()
	for name, raw := range configs {
		if _, exists := r.plugins[name]; !exists {
			r.mu.RUnlock()
			return fmt.Errorf("unknown plugin %s", name)
		}
		if !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			active[name] = raw
		}
	}
	r.mu.RUnlock()
	return r.Validate(active)
}

// Terminal reports whether the effective configuration has a response-producing
// plugin. It checks registration metadata only; parameter validation is done by
// Validate or Build. At most one terminal plugin may be active on a route.
func (r *Registry) Terminal(configs map[string]json.RawMessage) (bool, error) {
	name, err := r.TerminalName(configs)
	return name != "", err
}

// TerminalName identifies the effective response handler for route inspection.
func (r *Registry) TerminalName(configs map[string]json.RawMessage) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var terminal string
	for name, raw := range configs {
		entry, exists := r.plugins[name]
		if !exists {
			return "", fmt.Errorf("unknown plugin %s", name)
		}
		if entry.terminal && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			if terminal != "" {
				return "", fmt.Errorf("a route may configure only one terminal plugin")
			}
			terminal = name
		}
	}
	return terminal, nil
}

// Build constructs state once. The returned middleware may wrap several
// handlers, preserving shared counters, and is safe for concurrent requests.
func (r *Registry) Build(configs map[string]json.RawMessage) (Middleware, error) {
	if _, err := r.Terminal(configs); err != nil {
		return nil, err
	}
	type selected struct {
		name string
		registration
	}
	entries := make([]selected, 0, len(configs))
	r.mu.RLock()
	for name := range configs {
		entry, exists := r.plugins[name]
		if !exists {
			r.mu.RUnlock()
			return nil, fmt.Errorf("unknown plugin %s", name)
		}
		entries = append(entries, selected{name, entry})
	}
	r.mu.RUnlock()
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].priority == entries[j].priority {
			return entries[i].name < entries[j].name
		}
		return entries[i].priority > entries[j].priority
	})
	chain := make([]Middleware, 0, len(entries))
	for _, entry := range entries {
		if bytes.Equal(bytes.TrimSpace(configs[entry.name]), []byte("null")) {
			return nil, fmt.Errorf("plugin %s: null is only valid in an inheritance layer", entry.name)
		}
		middleware, err := entry.factory(configs[entry.name])
		if err != nil {
			return nil, fmt.Errorf("plugin %s: %w", entry.name, err)
		}
		if middleware == nil {
			return nil, fmt.Errorf("plugin %s returned no middleware", entry.name)
		}
		chain = append(chain, middleware)
	}
	return func(next http.Handler) http.Handler {
		for i := len(chain) - 1; i >= 0; i-- {
			next = chain[i](next)
		}
		return next
	}, nil
}

func decode(raw json.RawMessage, target any) error {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) < 2 || trimmed[0] != '{' {
		return fmt.Errorf("configuration must be an object")
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("configuration must contain one object")
	}
	return nil
}

func respond(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": status, "message": message})
	}
}
