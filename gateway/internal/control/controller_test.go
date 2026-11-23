package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"sitehub/gateway/internal/engine"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

type memoryStore struct {
	mu        sync.Mutex
	record    Record
	beforeCAS func()
	writeErr  error
	writes    int
}

func (s *memoryStore) Read(context.Context) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Record{Revision: s.record.Revision, Data: bytes.Clone(s.record.Data)}, nil
}

func (s *memoryStore) CompareAndSwap(_ context.Context, expected int64, data []byte) (Record, error) {
	if s.beforeCAS != nil {
		s.beforeCAS()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.writeErr != nil {
		return Record{}, s.writeErr
	}
	if s.record.Revision != expected {
		return Record{}, ErrConflict
	}
	s.record = Record{Revision: expected + 1, Data: bytes.Clone(data)}
	s.writes++
	return Record{Revision: s.record.Revision, Data: bytes.Clone(data)}, nil
}

const controlTestKey = "test-control-key-with-at-least-32-characters"

func staticDocument(version string) manifest.Document {
	config, _ := json.Marshal(map[string]any{"status": 200, "body": version})
	return manifest.Document{
		Version: 1,
		Routes: []manifest.Route{
			{ID: "first", Host: "gateway.test", Paths: []string{"/first"}, Plugins: manifest.Plugins{"static-response": config}},
			{ID: "second", Host: "gateway.test", Paths: []string{"/second"}, Plugins: manifest.Plugins{"static-response": config}},
		},
	}
}

func newTestController(t *testing.T, store Store) *Controller {
	t.Helper()
	allowed, err := AllowedCIDRs([]string{"172.31.0.0/16"})
	if err != nil {
		t.Fatal(err)
	}
	e := engine.New(nil, plugin.DefaultRegistry(), 4<<20)
	t.Cleanup(e.Close)
	c := &Controller{Store: store, Engine: e, Key: controlTestKey, Allowed: allowed}
	record, err := store.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if record.Revision > 0 {
		if err := c.Apply(record); err != nil {
			t.Fatal(err)
		}
	}
	return c
}

func adminRequest(c *Controller, method, path string, data []byte, etag string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewReader(data))
	r.RemoteAddr = "172.31.0.20:55000"
	r.Header.Set("Authorization", "Bearer "+controlTestKey)
	if etag != "" {
		r.Header.Set("If-Match", etag)
	}
	w := httptest.NewRecorder()
	c.Handler().ServeHTTP(w, r)
	return w
}

func assertLive(t *testing.T, c *Controller, want string) {
	t.Helper()
	for _, path := range []string{"/first", "/second"} {
		r := httptest.NewRequest(http.MethodGet, "http://gateway.test"+path, nil)
		w := httptest.NewRecorder()
		c.Engine.ServeHTTP(w, r)
		if w.Code != 200 || w.Body.String() != want {
			t.Fatalf("public %s = %d %q, want 200 %q", path, w.Code, w.Body.String(), want)
		}
	}
}

func TestAdminRequiresRealControlNetworkAndSingleBearerCredential(t *testing.T) {
	store := &memoryStore{record: Record{Revision: 1, Data: manifest.Bytes(staticDocument("old"))}}
	c := newTestController(t, store)
	for _, test := range []struct {
		name, remote, path, forwarded string
		auth                          []string
		want                          int
	}{
		{name: "allowed", remote: "172.31.0.20:123", path: "/v1/config", auth: []string{"Bearer " + controlTestKey}, want: 200},
		{name: "missing bearer", remote: "172.31.0.20:123", path: "/v1/config", want: 401},
		{name: "wrong bearer", remote: "172.31.0.20:123", path: "/v1/config", auth: []string{"Bearer incorrect"}, want: 401},
		{name: "duplicate bearer", remote: "172.31.0.20:123", path: "/v1/config", auth: []string{"Bearer " + controlTestKey, "Bearer " + controlTestKey}, want: 401},
		{name: "external with valid key", remote: "203.0.113.20:123", path: "/v1/config", auth: []string{"Bearer " + controlTestKey}, want: 403},
		{name: "forged forwarding", remote: "203.0.113.20:123", path: "/v1/config", forwarded: "172.31.0.20", auth: []string{"Bearer " + controlTestKey}, want: 403},
		{name: "invalid remote", remote: "172.31.0.20", path: "/v1/config", auth: []string{"Bearer " + controlTestKey}, want: 403},
		{name: "internal health", remote: "172.31.0.20:123", path: "/healthz", want: 200},
		{name: "external health", remote: "203.0.113.20:123", path: "/healthz", want: 403},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, test.path, nil)
			r.RemoteAddr = test.remote
			r.Header["Authorization"] = test.auth
			r.Header.Set("X-Forwarded-For", test.forwarded)
			w := httptest.NewRecorder()
			c.Handler().ServeHTTP(w, r)
			if w.Code != test.want {
				t.Fatalf("status = %d, want %d: %s", w.Code, test.want, w.Body.String())
			}
		})
	}
	r := httptest.NewRequest(http.MethodGet, "http://gateway.test/v1/config", nil)
	r.Header.Set("Authorization", "Bearer "+controlTestKey)
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatal("public handler exposed the separate configuration API")
	}
}

func TestGetExposesRevisionAndPublishingRequiresIfMatch(t *testing.T) {
	store := &memoryStore{record: Record{Revision: 7, Data: manifest.Bytes(staticDocument("old"))}}
	c := newTestController(t, store)
	w := adminRequest(c, http.MethodGet, "/v1/config", nil, "")
	var state State
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if w.Code != 200 || w.Header().Get("ETag") != `"7"` || state.Revision != 7 || state.AppliedRevision != 7 || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("GET revision contract: %d %v %s", w.Code, w.Header(), w.Body.String())
	}
	for _, etag := range []string{"", "*", `"nonsense"`, `"0"`, `"-1"`} {
		w = adminRequest(c, http.MethodPut, "/v1/config", manifest.Bytes(staticDocument("new")), etag)
		if w.Code != http.StatusPreconditionRequired {
			t.Fatalf("If-Match %q accepted with status %d", etag, w.Code)
		}
	}
	assertLive(t, c, "old")
	if record, _ := store.Read(context.Background()); record.Revision != 7 || store.writes != 0 {
		t.Fatal("missing precondition changed persisted configuration")
	}
}

func TestInvalidOrStalePublishChangesNeitherLiveNorStoredConfiguration(t *testing.T) {
	unknownPlugin := staticDocument("new")
	unknownPlugin.Routes[1].Plugins["unregistered"] = json.RawMessage(`{}`)
	invalidPlugin := staticDocument("new")
	invalidPlugin.Routes[1].Plugins["static-response"] = json.RawMessage(`{"status":999,"body":"new"}`)
	unknownField := bytes.Replace(manifest.Bytes(staticDocument("new")), []byte(`"version":1`), []byte(`"version":1,"unexpected":true`), 1)
	for _, test := range []struct {
		name string
		data []byte
		etag string
		want int
	}{
		{"unknown plugin", manifest.Bytes(unknownPlugin), `"2"`, 400},
		{"invalid plugin", manifest.Bytes(invalidPlugin), `"2"`, 400},
		{"unknown field", unknownField, `"2"`, 400},
		{"stale CAS", manifest.Bytes(staticDocument("new")), `"1"`, 409},
	} {
		t.Run(test.name, func(t *testing.T) {
			original := manifest.Bytes(staticDocument("old"))
			store := &memoryStore{record: Record{Revision: 2, Data: original}}
			c := newTestController(t, store)
			w := adminRequest(c, http.MethodPut, "/v1/config", test.data, test.etag)
			if w.Code != test.want {
				t.Fatalf("status = %d, want %d: %s", w.Code, test.want, w.Body.String())
			}
			assertLive(t, c, "old")
			record, _ := store.Read(context.Background())
			if record.Revision != 2 || !bytes.Equal(record.Data, original) || store.writes != 0 {
				t.Fatal("rejected publish changed the persisted document")
			}
		})
	}
}

func TestPublishCommitsWholeDocumentOnlyAfterPersistenceSucceeds(t *testing.T) {
	entered, release := make(chan struct{}, 1), make(chan struct{})
	store := &memoryStore{record: Record{Revision: 1, Data: manifest.Bytes(staticDocument("old"))}, beforeCAS: func() { entered <- struct{}{}; <-release }}
	c := newTestController(t, store)
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		result <- adminRequest(c, http.MethodPut, "/v1/config", manifest.Bytes(staticDocument("new")), `"1"`)
	}()
	<-entered
	assertLive(t, c, "old")
	before, _ := store.Read(context.Background())
	if before.Revision != 1 {
		t.Fatal("configuration changed before commit")
	}
	close(release)
	w := <-result
	if w.Code != 200 || w.Header().Get("ETag") != `"2"` {
		t.Fatalf("publish failed: %d %s", w.Code, w.Body.String())
	}
	assertLive(t, c, "new")
	after, _ := store.Read(context.Background())
	if after.Revision != 2 || !bytes.Equal(after.Data, manifest.Bytes(staticDocument("new"))) {
		t.Fatal("publication did not persist the complete document")
	}
	store.beforeCAS = nil
	store.writeErr = errors.New("store unavailable")
	w = adminRequest(c, http.MethodPut, "/v1/config", manifest.Bytes(staticDocument("uncommitted")), `"2"`)
	if w.Code != 503 {
		t.Fatalf("store failure status = %d", w.Code)
	}
	assertLive(t, c, "new")
}

func TestTwoGatewayControllersCannotBothPublishTheSameRevision(t *testing.T) {
	entered, release := make(chan struct{}, 2), make(chan struct{})
	store := &memoryStore{record: Record{Revision: 1, Data: manifest.Bytes(staticDocument("old"))}, beforeCAS: func() { entered <- struct{}{}; <-release }}
	controllers := []*Controller{newTestController(t, store), newTestController(t, store)}
	type result struct {
		index int
		w     *httptest.ResponseRecorder
	}
	results := make(chan result, 2)
	for i, c := range controllers {
		go func(index int, controller *Controller) {
			results <- result{index, adminRequest(controller, http.MethodPut, "/v1/config", manifest.Bytes(staticDocument("writer-"+strconv.Itoa(index))), `"1"`)}
		}(i, c)
	}
	<-entered
	<-entered
	close(release)
	winner, conflicts := -1, 0
	for range controllers {
		r := <-results
		switch r.w.Code {
		case 200:
			if winner >= 0 {
				t.Fatal("two writers committed the same expected revision")
			}
			winner = r.index
		case 409:
			conflicts++
		default:
			t.Fatalf("unexpected publish status %d: %s", r.w.Code, r.w.Body.String())
		}
	}
	if winner < 0 || conflicts != 1 || store.writes != 1 {
		t.Fatalf("winner=%d conflicts=%d writes=%d", winner, conflicts, store.writes)
	}
	latest, _ := store.Read(context.Background())
	for i, c := range controllers {
		if i == winner {
			assertLive(t, c, "writer-"+strconv.Itoa(winner))
		} else {
			assertLive(t, c, "old")
		}
		// Replicas converge through the authoritative watch after CAS succeeds.
		if err := c.Apply(latest); err != nil {
			t.Fatal(err)
		}
		assertLive(t, c, "writer-"+strconv.Itoa(winner))
	}
}

func TestApplyRejectsInvalidRevisionAndIgnoresOlderSnapshots(t *testing.T) {
	store := &memoryStore{record: Record{Revision: 1, Data: manifest.Bytes(staticDocument("first"))}}
	c := newTestController(t, store)
	if err := c.Apply(Record{Revision: 3, Data: manifest.Bytes(staticDocument("third"))}); err != nil {
		t.Fatal(err)
	}
	if err := c.Apply(Record{Revision: 2, Data: manifest.Bytes(staticDocument("second"))}); err != nil {
		t.Fatal(err)
	}
	assertLive(t, c, "third")
	invalid := staticDocument("fourth")
	invalid.Routes[1].Plugins["unknown"] = json.RawMessage(`{}`)
	if err := c.Apply(Record{Revision: 4, Data: manifest.Bytes(invalid)}); err == nil {
		t.Fatal("invalid watched document replaced the active handler")
	}
	assertLive(t, c, "third")
	w := adminRequest(c, http.MethodGet, "/v1/status", nil, "")
	var status struct {
		Applied int64  `json:"applied_revision"`
		Error   string `json:"last_error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.Applied != 3 || !strings.Contains(status.Error, "unknown") {
		t.Fatalf("rejected watch status = %+v", status)
	}
	if err := c.Apply(Record{Revision: 5, Data: manifest.Bytes(staticDocument("fifth"))}); err != nil {
		t.Fatal(err)
	}
	assertLive(t, c, "fifth")
	w = adminRequest(c, http.MethodGet, "/v1/status", nil, "")
	if strings.Contains(w.Body.String(), `"last_error":"unknown`) || !strings.Contains(w.Body.String(), `"last_error":""`) {
		t.Fatalf("recovery retained rejection: %s", w.Body.String())
	}
	c.Stop()
	if w = adminRequest(c, http.MethodGet, "/readyz", nil, ""); w.Code != 503 {
		t.Fatal("stopping controller remained ready")
	}
	if w = adminRequest(c, http.MethodPut, "/v1/config", manifest.Bytes(staticDocument("after-stop")), `"1"`); w.Code != 503 {
		t.Fatal("stopping controller accepted a publication")
	}
}

func TestInFlightPublishDuringStopReportsStoredOnly(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	store := &memoryStore{record: Record{Revision: 7, Data: manifest.Bytes(staticDocument("original"))}, beforeCAS: func() { close(entered); <-release }}
	c := newTestController(t, store)
	result := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		result <- adminRequest(c, http.MethodPut, "/v1/config", manifest.Bytes(staticDocument("next")), `"7"`)
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("publish did not reach CAS")
	}
	stopped := make(chan struct{})
	go func() { c.Stop(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		close(release)
		t.Fatal("stop blocked the shutdown deadline on an in-flight store transaction")
	}
	c.Engine.DrainStreams(t.Context())
	close(release)
	w := <-result
	var state State
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil || w.Code != http.StatusAccepted || state.Revision != 8 || state.AppliedRevision != 7 || state.LastError == "" {
		t.Fatalf("stored-only outcome = HTTP %d %s", w.Code, w.Body.String())
	}
	stored, _ := store.Read(t.Context())
	if err := c.Apply(stored); err == nil || c.applied != 7 {
		t.Fatal("stopping watcher published a new revision")
	}
	assertLive(t, c, "original")
}

// Keep blocked helper calls bounded when a regression prevents a result.
func awaitRecord(t *testing.T, ch <-chan Record) Record {
	t.Helper()
	select {
	case record := <-ch:
		return record
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for store operation")
		return Record{}
	}
}
