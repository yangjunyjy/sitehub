package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func build(t *testing.T, config string, next http.Handler) http.Handler {
	t.Helper()
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(config), &values); err != nil {
		t.Fatal(err)
	}
	middleware, err := DefaultRegistry().Build(values)
	if err != nil {
		t.Fatal(err)
	}
	return middleware(next)
}

func TestRegistryValidationAndOrder(t *testing.T) {
	r := DefaultRegistry()
	for _, item := range []string{
		`{"unknown":{}}`, `{"request-id":{"typo":true}}`, `{"request-id":null}`,
		`{"rate-limit":{"rate":0,"burst":1}}`, `{"limit-count":{"count":1,"window_seconds":0}}`,
		`{"browser-session":{"audience":"admin","origins":["https://admin.example/path"]}}`,
		`{"circuit-breaker":{"initial_open_seconds":31,"max_open_seconds":30}}`,
		`{"static-response":{"status":404,"headers":{"Connection":"close"}}}`,
	} {
		var values map[string]json.RawMessage
		_ = json.Unmarshal([]byte(item), &values)
		if err := r.Validate(values); err == nil {
			t.Errorf("accepted %s", item)
		}
	}
	registry := NewRegistry()
	var order []string
	for _, item := range []struct {
		name     string
		priority int
	}{{"a", 10}, {"b", 20}, {"c", 20}} {
		name := item.name
		err := registry.Register(name, item.priority, func(json.RawMessage) (Middleware, error) {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { order = append(order, name); next.ServeHTTP(w, r) })
			}, nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := registry.Register("a", 10, requestID); err == nil {
		t.Fatal("duplicate registration accepted")
	}
	middleware, err := registry.Build(map[string]json.RawMessage{"a": []byte(`{}`), "b": []byte(`{}`), "c": []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if strings.Join(order, ",") != "b,c,a" {
		t.Fatalf("unexpected order %v", order)
	}
}

func TestBrowserSessionIsolationAndOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"authorization": r.Header.Get("Authorization"), "cookie": r.Header.Get("Cookie"), "requestId": r.Header.Get("X-Request-ID")})
	})
	handler := build(t, `{"request-id":{},"browser-session":{"audience":"admin","origins":["https://admin.example"]}}`, next)
	server := httptest.NewServer(handler)
	defer server.Close()
	tests := []struct {
		name, method, path, cookie, auth, origin string
		status                                   int
		want                                     string
	}{
		{"admin-cookie", "GET", "/v1/rbac/users", "sitehub_admin_session=admin; sitehub_site_session=member", "", "", 200, "Bearer admin"},
		{"member-isolation", "GET", "/v1/rbac/users", "sitehub_site_session=member", "", "", 200, ""},
		{"explicit-bearer", "POST", "/v1/rbac/users", "sitehub_admin_session=admin", "Bearer explicit", "", 200, "Bearer explicit"},
		{"cookie-write-missing-origin", "POST", "/v1/rbac/users", "sitehub_admin_session=admin", "", "", 403, ""},
		{"cookie-write-foreign-origin", "POST", "/v1/rbac/users", "sitehub_admin_session=admin", "", "https://foreign.example", 403, ""},
		{"cookie-write-approved-origin", "POST", "/v1/rbac/users", "sitehub_admin_session=admin", "", "https://admin.example", 200, "Bearer admin"},
		{"foreign-login", "POST", "/v1/rbac/auth/login", "", "Bearer explicit", "https://foreign.example", 403, ""},
		{"foreign-register", "POST", "/v1/rbac/auth/register", "", "", "https://foreign.example", 403, ""},
		{"cli-login", "POST", "/v1/rbac/auth/login", "", "", "", 200, ""},
		{"duplicate-cookie", "GET", "/", "sitehub_admin_session=a; sitehub_admin_session=b", "", "", 400, ""},
		{"duplicate-empty-cookie", "GET", "/", "sitehub_admin_session=; sitehub_admin_session=b", "", "", 400, ""},
		{"invalid-cookie", "GET", "/", "sitehub_admin_session=bad token", "", "", 400, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, server.URL+test.path, nil)
			if test.cookie != "" {
				req.Header.Set("Cookie", test.cookie)
			}
			if test.auth != "" {
				req.Header.Set("Authorization", test.auth)
			}
			if test.origin != "" {
				req.Header.Set("Origin", test.origin)
			}
			response, err := server.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status {
				t.Fatalf("status %d, want %d", response.StatusCode, test.status)
			}
			if response.Header.Get("X-Request-ID") == "" {
				t.Fatal("request ID absent on response")
			}
			if test.status == 200 {
				var result map[string]string
				if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}
				if result["authorization"] != test.want || result["cookie"] != "" || result["requestId"] != response.Header.Get("X-Request-ID") {
					t.Fatalf("unexpected upstream credentials or request ID: %v", result)
				}
			}
		})
	}
	for _, header := range []string{"Authorization", "Cookie", "Origin"} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Add(header, "one")
		req.Header.Add(header, "two")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != 400 {
			t.Fatalf("duplicate %s accepted", header)
		}
	}
}

func TestRealHTTPRejectsMixedCaseDuplicateCredentialHeaders(t *testing.T) {
	var reached atomic.Int64
	server := httptest.NewServer(build(t, `{"browser-session":{"audience":"admin","origins":["https://admin.example"]}}`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached.Add(1)
		w.WriteHeader(http.StatusNoContent)
	})))
	defer server.Close()
	address := strings.TrimPrefix(server.URL, "http://")
	for _, headers := range []string{
		"authorization: Bearer one\r\nAuthorization: Bearer two\r\n",
		"cookie: sitehub_admin_session=one\r\nCookie: sitehub_admin_session=two\r\n",
		"origin: https://admin.example\r\nOrigin: https://admin.example\r\n",
	} {
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
		_, _ = fmt.Fprintf(conn, "GET /v1/rbac/users HTTP/1.1\r\nHost: %s\r\n%sConnection: close\r\n\r\n", address, headers)
		response, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "GET"})
		if err != nil {
			_ = conn.Close()
			t.Fatal(err)
		}
		_, _ = io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		_ = conn.Close()
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("mixed-case duplicate headers accepted: %d", response.StatusCode)
		}
	}
	if reached.Load() != 0 {
		t.Fatal("ambiguous credentials reached upstream")
	}
}

func TestLimitsBoundedAndExpireWithoutResettingActiveClients(t *testing.T) {
	now := time.Now()
	rateState := &tokenLimits{entries: map[string]*tokenEntry{}, config: rateConfig{Rate: 1, Burst: 1, IdleTTLSeconds: 2, MaxKeys: 1}}
	if !rateState.allow("a", now) || rateState.allow("a", now) || rateState.allow("b", now) {
		t.Fatal("rate burst/capacity enforcement failed")
	}
	if rateState.allow("a", now.Add(time.Millisecond)) {
		t.Fatal("another client reset active bucket")
	}
	if !rateState.allow("b", now.Add(3*time.Second)) || len(rateState.entries) != 1 {
		t.Fatal("idle rate entry did not expire")
	}
	counts := &countLimits{entries: map[string]countEntry{}, config: countConfig{Count: 1, WindowSeconds: 2, MaxKeys: 1}}
	if ok, _ := counts.allow("a", now); !ok {
		t.Fatal("first request rejected")
	}
	if ok, _ := counts.allow("a", now); ok {
		t.Fatal("quota exceeded")
	}
	if ok, _ := counts.allow("b", now); ok {
		t.Fatal("capacity exceeded")
	}
	if ok, _ := counts.allow("b", now.Add(3*time.Second)); !ok || len(counts.entries) != 1 {
		t.Fatal("expired count did not release capacity")
	}
}

func TestConcurrentHTTPLimitsIgnoreForwardedIP(t *testing.T) {
	for _, config := range []string{`{"rate-limit":{"rate":0.01,"burst":5,"idle_ttl_seconds":600}}`, `{"limit-count":{"count":5,"window_seconds":60}}`} {
		server := httptest.NewServer(build(t, config, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })))
		var allowed, rejected atomic.Int64
		var wg sync.WaitGroup
		for i := 0; i < 30; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				req, _ := http.NewRequest("GET", server.URL, nil)
				req.Header.Set("X-Forwarded-For", fmt.Sprintf("10.0.0.%d", i))
				res, err := server.Client().Do(req)
				if err != nil {
					t.Error(err)
					return
				}
				defer res.Body.Close()
				_, _ = io.Copy(io.Discard, res.Body)
				switch res.StatusCode {
				case 204:
					allowed.Add(1)
				case 429:
					rejected.Add(1)
					if retry, err := strconv.Atoi(res.Header.Get("Retry-After")); err != nil || retry < 1 {
						t.Errorf("invalid Retry-After: %q", res.Header.Get("Retry-After"))
					}
				default:
					t.Errorf("unexpected status %d", res.StatusCode)
				}
			}(i)
		}
		wg.Wait()
		server.Close()
		if allowed.Load() != 5 || rejected.Load() != 25 {
			t.Fatalf("allowed=%d rejected=%d", allowed.Load(), rejected.Load())
		}
	}
}

func TestCircuitTransitionsSingleProbeAndStaleResponse(t *testing.T) {
	b := &breaker{config: breakerConfig{Failures: 2, Successes: 2, InitialOpenSeconds: 1, MaxOpenSeconds: 2}, state: "closed"}
	now := time.Now()
	old, _, _ := b.allow(now)
	b.finish(old, 500, now)
	b.finish(old, 502, now)
	if _, ok, _ := b.allow(now); ok {
		t.Fatal("breaker did not open")
	}
	b.finish(old, 200, now)
	if _, ok, _ := b.allow(now); ok {
		t.Fatal("stale success closed breaker")
	}
	probe, ok, _ := b.allow(now.Add(time.Second))
	if !ok {
		t.Fatal("half-open probe rejected")
	}
	if _, ok, _ := b.allow(now.Add(time.Second)); ok {
		t.Fatal("concurrent half-open probe admitted")
	}
	b.finish(probe, 503, now.Add(time.Second))
	if b.backoff != 2*time.Second {
		t.Fatal("failed probe did not increase backoff")
	}
	probe, ok, _ = b.allow(now.Add(3 * time.Second))
	if !ok {
		t.Fatal("second half-open probe rejected")
	}
	b.finish(probe, 206, now.Add(3*time.Second))
	probe, ok, _ = b.allow(now.Add(3 * time.Second))
	if !ok {
		t.Fatal("next recovery probe rejected")
	}
	b.finish(probe, 304, now.Add(3*time.Second))
	if b.state != "closed" || b.backoff != 0 {
		t.Fatal("successful probes did not close breaker")
	}
	b.finish(old, 500, now.Add(4*time.Second))
	b.finish(old, 500, now.Add(4*time.Second))
	if b.state != "closed" || b.failures != 0 {
		t.Fatal("failures from old generation reopened recovered breaker")
	}
}

func TestCircuitPreservesFailuresAndSharesStateAcrossWrappedHandlers(t *testing.T) {
	middleware, err := DefaultRegistry().Build(map[string]json.RawMessage{"circuit-breaker": []byte(`{"failures":2}`)})
	if err != nil {
		t.Fatal(err)
	}
	var reached atomic.Int64
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached.Add(1)
		w.WriteHeader(500)
		_, _ = w.Write([]byte("upstream error"))
	})
	server := httptest.NewServer(middleware(handler))
	defer server.Close()
	for _, expected := range []int{500, 500, 503} {
		response, err := server.Client().Get(server.URL)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != expected {
			t.Fatalf("status %d, want %d", response.StatusCode, expected)
		}
		if expected == 500 && string(data) != "upstream error" {
			t.Fatalf("upstream body changed: %s", data)
		}
	}
	other := httptest.NewRecorder()
	middleware(handler).ServeHTTP(other, httptest.NewRequest("GET", "/", nil))
	if other.Code != 503 || reached.Load() != 2 {
		t.Fatal("wrapping a new handler lost circuit state")
	}
}

func TestProxyStreamingRangeHeadAndUpgradeThroughCircuit(t *testing.T) {
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	defer unblock()
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/stream":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("first\n"))
			w.(http.Flusher).Flush()
			<-release
			_, _ = w.Write([]byte("second\n"))
		case "/upgrade":
			conn, rw, err := w.(http.Hijacker).Hijack()
			if err != nil {
				return
			}
			defer conn.Close()
			_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")
			_ = rw.Flush()
			line, _ := rw.ReadString('\n')
			_, _ = rw.WriteString(line)
			_ = rw.Flush()
		default:
			http.ServeContent(w, r, "file", time.Time{}, strings.NewReader("0123456789"))
		}
	}))
	defer backend.Close()
	u, _ := url.Parse(backend.URL)
	server := httptest.NewServer(build(t, `{"circuit-breaker":{}}`, httputil.NewSingleHostReverseProxy(u)))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL+"/stream", nil)
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(response.Body)
	first, err := reader.ReadString('\n')
	if err != nil || first != "first\n" {
		t.Fatalf("first stream chunk unavailable: %q %v", first, err)
	}
	unblock()
	second, err := reader.ReadString('\n')
	response.Body.Close()
	if err != nil || second != "second\n" {
		t.Fatalf("second stream chunk unavailable: %q %v", second, err)
	}
	for _, method := range []string{"GET", "HEAD"} {
		req, _ := http.NewRequest(method, server.URL+"/file", nil)
		req.Header.Set("Range", "bytes=2-5")
		response, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != 206 || response.Header.Get("Content-Range") != "bytes 2-5/10" {
			t.Fatalf("range response changed: %d %v", response.StatusCode, response.Header)
		}
		if method == "GET" && string(data) != "2345" || method == "HEAD" && len(data) != 0 {
			t.Fatalf("invalid %s range body: %q", method, data)
		}
	}
	address := strings.TrimPrefix(server.URL, "http://")
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, _ = fmt.Fprintf(conn, "GET /upgrade HTTP/1.1\r\nHost: %s\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n", address)
	buffer := bufio.NewReader(conn)
	res, err := http.ReadResponse(buffer, &http.Request{Method: "GET"})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 101 {
		t.Fatalf("upgrade failed: %d", res.StatusCode)
	}
	_, _ = conn.Write([]byte("upgrade-echo\n"))
	echo, err := buffer.ReadString('\n')
	if err != nil || echo != "upgrade-echo\n" {
		t.Fatalf("upgrade echo failed: %q %v", echo, err)
	}
}

func TestStaticAndFrontEndStrip(t *testing.T) {
	server := httptest.NewServer(build(t, `{"request-id":{},"static-response":{"status":404,"body":"missing","headers":{"Content-Type":"text/plain"}}}`, nil))
	defer server.Close()
	for _, method := range []string{"GET", "HEAD"} {
		req, _ := http.NewRequest(method, server.URL, nil)
		response, err := server.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != 404 || response.ContentLength != 7 || response.Header.Get("X-Request-ID") == "" {
			t.Fatal("static response metadata lost")
		}
		if method == "HEAD" && len(data) != 0 {
			t.Fatal("HEAD returned body")
		}
	}
	request := httptest.NewRequest("GET", "/", nil)
	for _, name := range []string{"Cookie", "Authorization", "X-Service-Key", "X-Actor-ID", "X-User-ID", "X-Role", "X-Roles"} {
		request.Header.Set(name, "secret")
	}
	build(t, `{"strip-credentials":{}}`, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header) != 0 {
			t.Fatalf("credentials reached frontend: %v", r.Header)
		}
	})).ServeHTTP(httptest.NewRecorder(), request)
}
