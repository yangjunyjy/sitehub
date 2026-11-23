package control

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"sitehub/gateway/diagnostics"
	"sitehub/gateway/internal/engine"
	"sitehub/gateway/manifest"
)

type State = diagnostics.State

var generationPattern = regexp.MustCompile("^[0-9a-f]{24}$")

type Controller struct {
	HTTPListen string
	Store      Store
	Engine     *engine.Engine
	Key        string
	Allowed    []*net.IPNet
	// Load is set only in file mode. PUT is then disabled so restarts and reloads
	// always use the same authoritative files, including their include tree.
	Load      func() (manifest.Document, error)
	mu        sync.Mutex
	applied   int64
	lastError string
	stopping  atomic.Bool
}

func (c *Controller) Apply(record Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopping.Load() {
		return errors.New("gateway is stopping")
	}
	if len(record.Data) == 0 {
		c.lastError = "active configuration missing; retaining applied snapshot"
		return errors.New(c.lastError)
	}
	if record.Revision <= c.applied {
		if record.Revision == c.applied {
			c.lastError = ""
		}
		return nil
	}
	d, err := manifest.Decode(record.Data)
	if err != nil {
		c.lastError = err.Error()
		return err
	}
	prepared, err := c.Engine.Prepare(d)
	if err != nil {
		c.lastError = err.Error()
		return err
	}
	return c.install(prepared, record.Revision)
}

// install runs under c.mu. A store transaction may finish after shutdown
// starts. Only advance applied when the engine actually installed its snapshot.
func (c *Controller) install(prepared *engine.Snapshot, revision int64) error {
	if c.stopping.Load() {
		prepared.Release()
	} else if c.Engine.Commit(prepared) {
		c.applied = revision
		c.lastError = ""
		return nil
	}
	c.lastError = "gateway is stopping; stored configuration was not applied locally"
	return errors.New(c.lastError)
}
func (c *Controller) Failure(err error) { c.mu.Lock(); defer c.mu.Unlock(); c.lastError = err.Error() }
func (c *Controller) Stop() {
	c.stopping.Store(true)
}
func (c *Controller) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { respond(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if c.stopping.Load() || c.Engine.Revision() == "" {
			respond(w, 503, map[string]string{"status": "not ready"})
			return
		}
		respond(w, 200, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/v1/config", c.config)
	mux.HandleFunc("GET /v1/releases/{generation}", c.receipt)
	mux.HandleFunc("POST /v1/reload", c.reload)
	mux.HandleFunc("POST /v1/match", c.match)
	mux.Handle("GET /metrics", c.Engine.Metrics())
	mux.HandleFunc("GET /v1/status", func(w http.ResponseWriter, _ *http.Request) {
		c.mu.Lock()
		defer c.mu.Unlock()
		respond(w, 200, map[string]any{"applied_revision": c.applied, "content_hash": c.Engine.Revision(), "last_error": c.lastError, "stopping": c.stopping.Load()})
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		allowed := false
		if err == nil {
			for _, network := range c.Allowed {
				if network.Contains(net.ParseIP(ip)) {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			respond(w, 403, map[string]string{"message": "Control network required"})
			return
		}
		if r.URL.Path != "/healthz" && r.URL.Path != "/readyz" {
			auth := r.Header.Values("Authorization")
			if len(auth) != 1 || subtle.ConstantTimeCompare([]byte(auth[0]), []byte("Bearer "+c.Key)) != 1 {
				respond(w, 401, map[string]string{"message": "Control authentication required"})
				return
			}
		}
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})
}
func (c *Controller) receipt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("generation")
	if !generationPattern.MatchString(id) {
		respond(w, 400, map[string]string{"message": "Invalid release ID"})
		return
	}
	store, ok := c.Store.(receiptStore)
	if !ok {
		respond(w, 404, map[string]string{"message": "Release receipt unavailable"})
		return
	}
	record, err := store.Receipt(r.Context(), id)
	if err != nil {
		respond(w, 503, map[string]string{"message": err.Error()})
		return
	}
	if record.Revision == 0 {
		respond(w, 404, map[string]string{"message": "Release receipt not found"})
		return
	}
	respond(w, 200, map[string]any{"revision": record.Revision, "receipt": string(record.Data)})
}

func (c *Controller) config(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		record, err := c.Store.Read(r.Context())
		var d manifest.Document
		if err == nil {
			d, err = manifest.Decode(record.Data)
			if err != nil {
				err = errors.New("Stored configuration invalid or missing")
			}
		}
		if err != nil && r.URL.Query().Get("local_fallback") != "1" {
			respond(w, 503, map[string]string{"message": err.Error()})
			return
		}
		c.mu.Lock()
		state := State{FileManaged: c.Load != nil, Revision: record.Revision, AppliedRevision: c.applied, Manifest: d, LastError: c.lastError}
		state.Runtime, state.AppliedManifest = c.Engine.Runtime(c.HTTPListen)
		c.mu.Unlock()
		if err != nil {
			if state.AppliedManifest == nil {
				respond(w, 503, map[string]string{"message": err.Error()})
				return
			}
			state.Revision = 0
			state.DesiredUnavailable = true
			state.Manifest = *state.AppliedManifest
			state.LastError = "无法读取期望配置：" + err.Error()
			respond(w, 200, state)
			return
		}
		if record.Revision == state.AppliedRevision {
			state.AppliedManifest = nil
		}
		w.Header().Set("ETag", strconv.Quote(strconv.FormatInt(record.Revision, 10)))
		respond(w, 200, state)
		return
	}
	if r.Method != http.MethodPut {
		w.Header().Set("Allow", "GET, PUT")
		respond(w, 405, map[string]string{"message": "Method not allowed"})
		return
	}
	if c.Load != nil {
		respond(w, 405, map[string]string{"message": "Configuration is managed by files; edit them and use POST /v1/reload"})
		return
	}
	if c.stopping.Load() {
		respond(w, 503, map[string]string{"message": "Gateway is stopping", "outcome": "rejected"})
		return
	}
	expected, err := strconv.ParseInt(strings.Trim(r.Header.Get("If-Match"), "\""), 10, 64)
	if err != nil || expected < 1 {
		respond(w, 428, map[string]string{"message": "If-Match with the current revision is required"})
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4<<20))
	if err != nil {
		respond(w, 413, map[string]string{"message": "Manifest exceeds size limit"})
		return
	}
	d, err := manifest.Decode(data)
	if err != nil {
		respond(w, 400, map[string]string{"message": err.Error()})
		return
	}
	// Serialize local watch and publish. Compare-and-swap in etcd serializes
	// writers across gateway replicas; only complete validated snapshots commit.
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopping.Load() {
		respond(w, 503, map[string]string{"message": "Gateway is stopping", "outcome": "rejected"})
		return
	}
	generation := r.Header.Get("X-Release-ID")
	store, supportsReceipts := c.Store.(receiptStore)
	if generation != "" && (!generationPattern.MatchString(generation) || !supportsReceipts) {
		respond(w, 400, map[string]string{"message": "Invalid release ID or store does not support receipts"})
		return
	}
	prepared, err := c.Engine.Prepare(d)
	if err != nil {
		respond(w, 400, map[string]string{"message": err.Error()})
		return
	}
	var record Record
	if generation != "" {
		record, err = store.PublishRelease(r.Context(), expected, manifest.Bytes(d), generation)
	} else {
		record, err = c.Store.CompareAndSwap(r.Context(), expected, manifest.Bytes(d))
	}
	if err != nil {
		prepared.Release()
		code := 503
		if errors.Is(err, ErrConflict) {
			code = 409
		}
		respond(w, code, map[string]string{"message": err.Error()})
		return
	}
	w.Header().Set("ETag", strconv.Quote(strconv.FormatInt(record.Revision, 10)))
	if err := c.install(prepared, record.Revision); err != nil {
		respond(w, http.StatusAccepted, State{Revision: record.Revision, AppliedRevision: c.applied, Manifest: d, LastError: err.Error()})
		return
	}
	respond(w, 200, State{Revision: record.Revision, AppliedRevision: c.applied, Manifest: d})
}

func (c *Controller) reload(w http.ResponseWriter, r *http.Request) {
	if c.Load == nil {
		respond(w, 409, map[string]string{"message": "Reload is available in file source mode; use configuration publication for etcd"})
		return
	}
	if c.stopping.Load() {
		respond(w, 503, map[string]string{"message": "Gateway is stopping"})
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stopping.Load() {
		respond(w, 503, map[string]string{"message": "Gateway is stopping"})
		return
	}
	d, err := c.Load()
	if err != nil {
		c.lastError = err.Error()
		respond(w, 400, map[string]string{"message": err.Error()})
		return
	}
	if manifest.Revision(d) == c.Engine.Revision() {
		c.lastError = ""
		respond(w, 200, State{Revision: c.applied, AppliedRevision: c.applied, Manifest: d})
		return
	}
	prepared, err := c.Engine.Prepare(d)
	if err != nil {
		c.lastError = err.Error()
		respond(w, 400, map[string]string{"message": err.Error()})
		return
	}
	record, err := c.Store.CompareAndSwap(r.Context(), c.applied, manifest.Bytes(d))
	if err != nil {
		prepared.Release()
		c.lastError = err.Error()
		respond(w, 503, map[string]string{"message": err.Error()})
		return
	}
	if err := c.install(prepared, record.Revision); err != nil {
		respond(w, 503, map[string]string{"message": err.Error()})
		return
	}
	respond(w, 200, State{Revision: record.Revision, AppliedRevision: c.applied, Manifest: d})
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func Seed(ctx context.Context, store Store, d manifest.Document) (Record, error) {
	record, err := store.Read(ctx)
	if err != nil {
		return record, err
	}
	if record.Revision != 0 {
		return record, nil
	}
	record, err = store.CompareAndSwap(ctx, 0, manifest.Bytes(d))
	if errors.Is(err, ErrConflict) {
		return store.Read(ctx)
	}
	return record, err
}
func AllowedCIDRs(values []string) ([]*net.IPNet, error) {
	var result []*net.IPNet
	for _, value := range values {
		_, network, err := net.ParseCIDR(value)
		if err != nil {
			return nil, fmt.Errorf("invalid Admin.AllowedCIDRs")
		}
		result = append(result, network)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("Admin.AllowedCIDRs cannot be empty")
	}
	return result, nil
}
