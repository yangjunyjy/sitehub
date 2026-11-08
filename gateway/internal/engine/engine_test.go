package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

func testDocument(target string) manifest.Document {
	return manifest.Document{
		Version: 1,
		Upstreams: map[string]manifest.Upstream{"app": {
			Nodes: []manifest.Node{{URL: target, Weight: 1}}, ConnectTimeoutMS: 1000,
			ResponseHeaderTimeoutMS: 2000, IOTimeoutMS: 2000,
			HealthPath: "/_ready", HealthIntervalMS: 100, HealthTimeoutMS: 300, HealthFailures: 1,
		}},
		Profiles: map[string]manifest.Plugins{},
		Routes:   []manifest.Route{{ID: "app", Host: "example.test", Paths: []string{"/*"}, Upstream: "app", Plugins: manifest.Plugins{"request-id": []byte(`{}`)}}},
	}
}

func testBackend(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_ready" {
			w.WriteHeader(204)
			return
		}
		handler.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	return server
}

func commit(t *testing.T, e *Engine, d manifest.Document) {
	t.Helper()
	s, err := e.Prepare(d)
	if err != nil {
		t.Fatal(err)
	}
	e.Commit(s)
}

func eventually(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out: " + description)
}

func countHealthy(e *Engine, upstream string) int {
	s := e.current.Load()
	if s == nil {
		return 0
	}
	p := s.pools[upstream]
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	count := 0
	for _, node := range p.nodes {
		if node.healthy {
			count++
		}
	}
	return count
}

func testGateway(t *testing.T, d manifest.Document, maxBody int64) (*Engine, *httptest.Server) {
	t.Helper()
	e := New(nil, plugin.DefaultRegistry(), maxBody)
	t.Cleanup(e.Close)
	commit(t, e, d)
	server := httptest.NewServer(e)
	t.Cleanup(server.Close)
	for name, u := range d.Upstreams {
		eventually(t, "upstream ready", func() bool { return countHealthy(e, name) == len(u.Nodes) })
	}
	return e, server
}

func gatewayRequest(t *testing.T, server *httptest.Server, method, path string, body io.Reader, headers http.Header) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "example.test"
	if headers != nil {
		req.Header = headers.Clone()
	}
	res, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func readResponse(t *testing.T, res *http.Response) string {
	t.Helper()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestProxyRestoresValidatedHopHeadersAndSanitizesIdentity(t *testing.T) {
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(r.Header) }))
	d := testDocument(backend.URL)
	d.Routes[0].Plugins["browser-session"] = []byte(`{"audience":"admin","origins":["https://example.test"]}`)
	_, gateway := testGateway(t, d, 1<<20)
	for _, test := range []struct{ auth, cookie, origin, want string }{
		{"", "sitehub_admin_session=cookie-token; sitehub_site_session=other", "https://example.test", "Bearer cookie-token"},
		{"Bearer explicit", "sitehub_admin_session=cookie-token", "", "Bearer explicit"},
		{"", "sitehub_site_session=member-token", "", ""},
	} {
		headers := http.Header{"Cookie": {test.cookie}, "X-Request-Id": {"fixed-request-id"}, "Connection": {"Authorization, X-Request-ID, Cookie"}, "X-Forwarded-For": {"203.0.113.2"}, "X-Real-Ip": {"203.0.113.3"}, "Forwarded": {"for=forged"}, "X-Forwarded-Host": {"evil.example"}, "X-Forwarded-Proto": {"https"}, "X-Actor-Id": {"999"}, "X-Service-Key": {"forged"}, "X-Role": {"admin"}}
		if test.auth != "" {
			headers.Set("Authorization", test.auth)
		}
		if test.origin != "" {
			headers.Set("Origin", test.origin)
		}
		res := gatewayRequest(t, gateway, "POST", "/v1/users", strings.NewReader("{}"), headers)
		var received http.Header
		if err := json.NewDecoder(res.Body).Decode(&received); err != nil {
			res.Body.Close()
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 200 || received.Get("Authorization") != test.want || received.Get("X-Request-ID") != "fixed-request-id" {
			t.Fatalf("validated headers lost: status=%d headers=%v", res.StatusCode, received)
		}
		if res.Header.Get("X-Request-ID") != "fixed-request-id" {
			t.Fatal("response request ID lost")
		}
		for _, header := range []string{"Cookie", "Forwarded", "X-Actor-ID", "X-Service-Key", "X-Role"} {
			if received.Get(header) != "" {
				t.Errorf("untrusted %s reached upstream", header)
			}
		}
		if received.Get("X-Forwarded-For") != "127.0.0.1" || received.Get("X-Real-IP") != "127.0.0.1" || received.Get("X-Forwarded-Host") != "example.test" || received.Get("X-Forwarded-Proto") != "http" {
			t.Fatalf("forged forwarding metadata survived: %v", received)
		}
	}
}

func TestProxyPreservesRangeHeadQueryAndIdempotency(t *testing.T) {
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Raw-Query", r.URL.RawQuery)
		w.Header().Set("X-Upstream-Idempotency", r.Header.Get("Idempotency-Key"))
		if r.Method == "POST" {
			_, _ = io.Copy(w, r.Body)
			return
		}
		http.ServeContent(w, r, "file.txt", time.Time{}, strings.NewReader("0123456789"))
	}))
	_, gateway := testGateway(t, testDocument(backend.URL), 1<<20)
	query := "q=a%2Bb&q=c+d&signed=%2Ftest%3D"
	for _, method := range []string{"GET", "HEAD"} {
		res := gatewayRequest(t, gateway, method, "/file?"+query, nil, http.Header{"Range": {"bytes=2-5"}})
		body := readResponse(t, res)
		if res.StatusCode != 206 || res.Header.Get("Content-Range") != "bytes 2-5/10" || res.Header.Get("X-Raw-Query") != query {
			t.Fatalf("range/query changed: %d %v", res.StatusCode, res.Header)
		}
		if method == "GET" && body != "2345" || method == "HEAD" && body != "" {
			t.Fatalf("wrong %s body: %q", method, body)
		}
	}
	res := gatewayRequest(t, gateway, "POST", "/mutation?"+query, strings.NewReader("payload"), http.Header{"Idempotency-Key": {"application-key"}})
	if body := readResponse(t, res); body != "payload" || res.Header.Get("X-Upstream-Idempotency") != "application-key" || res.Header.Get("X-Raw-Query") != query {
		t.Fatalf("mutation contract changed: %q %v", body, res.Header)
	}
}

func TestStreamSurvivesAtomicRouteReplacement(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("first\n"))
		w.(http.Flusher).Flush()
		<-release
		_, _ = w.Write([]byte("last\n"))
	}))
	d := testDocument(backend.URL)
	d.Routes[0].Plugins["circuit-breaker"] = []byte(`{}`)
	e, gateway := testGateway(t, d, 1<<20)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", gateway.URL+"/events", nil)
	req.Host = "example.test"
	res, err := gateway.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	reader := bufio.NewReader(res.Body)
	first, err := reader.ReadString('\n')
	if err != nil || first != "first\n" {
		t.Fatalf("stream buffered: %q %v", first, err)
	}
	d.Upstreams = map[string]manifest.Upstream{}
	d.Routes[0].Upstream = ""
	d.Routes[0].Plugins = manifest.Plugins{"static-response": []byte(`{"status":200,"body":"new-config"}`)}
	commit(t, e, d)
	if body := readResponse(t, gatewayRequest(t, gateway, "GET", "/events", nil, nil)); body != "new-config" {
		t.Fatalf("new requests missed replacement: %q", body)
	}
	unblock()
	tail, err := io.ReadAll(reader)
	if err != nil || string(tail) != "last\n" {
		t.Fatalf("in-flight stream truncated by reload: %q %v", tail, err)
	}
}

func TestInvalidCandidateAndUnchangedPluginsPreserveAppliedState(t *testing.T) {
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	d := testDocument(backend.URL)
	d.Routes[0].Plugins["limit-count"] = []byte(`{"count":2,"window_seconds":60}`)
	e, gateway := testGateway(t, d, 1<<20)
	res := gatewayRequest(t, gateway, "GET", "/first", nil, nil)
	readResponse(t, res)
	if res.StatusCode != 204 {
		t.Fatal(res.StatusCode)
	}
	original := e.Revision()
	bad, _ := manifest.Decode(manifest.Bytes(d))
	bad.Routes[0].Plugins["missing-plugin"] = []byte(`{}`)
	if prepared, err := e.Prepare(bad); err == nil {
		prepared.Release()
		t.Fatal("unknown plugin accepted")
	}
	if e.Revision() != original {
		t.Fatal("rejected configuration changed applied revision")
	}
	d.Routes = append(d.Routes, manifest.Route{ID: "new-route", Host: "other.example.test", Paths: []string{"/*"}, Plugins: manifest.Plugins{"static-response": []byte(`{"status":200,"body":"other"}`)}})
	commit(t, e, d)
	for _, want := range []int{204, 429} {
		res := gatewayRequest(t, gateway, "GET", "/after-reload", nil, nil)
		readResponse(t, res)
		if res.StatusCode != want {
			t.Fatalf("limiter reset or invalid count: %d want %d", res.StatusCode, want)
		}
	}
}

func TestHealthRemovalRecoveryAndWeightedBalancing(t *testing.T) {
	var firstHealthy atomic.Bool
	firstHealthy.Store(true)
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_ready" {
			if !firstHealthy.Load() {
				w.WriteHeader(503)
			} else {
				w.WriteHeader(204)
			}
			return
		}
		_, _ = w.Write([]byte("first"))
	}))
	t.Cleanup(first.Close)
	second := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("second")) }))
	d := testDocument(first.URL)
	u := d.Upstreams["app"]
	u.Nodes = append(u.Nodes, manifest.Node{URL: second.URL, Weight: 2})
	d.Upstreams["app"] = u
	e, gateway := testGateway(t, d, 1<<20)
	counts := map[string]int{}
	for i := 0; i < 12; i++ {
		counts[readResponse(t, gatewayRequest(t, gateway, "GET", "/data", nil, nil))]++
	}
	if counts["first"] != 4 || counts["second"] != 8 {
		t.Fatalf("weighted traffic distribution incorrect: %v", counts)
	}
	firstHealthy.Store(false)
	eventually(t, "unhealthy endpoint removed", func() bool { return countHealthy(e, "app") == 1 })
	for i := 0; i < 6; i++ {
		if got := readResponse(t, gatewayRequest(t, gateway, "GET", "/data", nil, nil)); got != "second" {
			t.Fatal("unhealthy endpoint selected")
		}
	}
	firstHealthy.Store(true)
	eventually(t, "endpoint recovery", func() bool { return countHealthy(e, "app") == 2 })
	counts = map[string]int{}
	for i := 0; i < 6; i++ {
		counts[readResponse(t, gatewayRequest(t, gateway, "GET", "/data", nil, nil))]++
	}
	if counts["first"] != 2 || counts["second"] != 4 {
		t.Fatalf("recovered endpoint missing: %v", counts)
	}
}

func TestIdleTimeoutStopsStalledBodyButAllowsActiveStream(t *testing.T) {
	stalledDone := make(chan struct{})
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "text/event-stream")
		if r.URL.Path == "/stalled" {
			_, _ = w.Write([]byte("first\n"))
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			close(stalledDone)
			return
		}
		for i := 0; i < 8; i++ {
			_, _ = fmt.Fprintf(w, "%d\n", i)
			w.(http.Flusher).Flush()
			time.Sleep(50 * time.Millisecond)
		}
	}))
	d := testDocument(backend.URL)
	u := d.Upstreams["app"]
	u.IOTimeoutMS = 200
	d.Upstreams["app"] = u
	_, gateway := testGateway(t, d, 1<<20)
	res := gatewayRequest(t, gateway, "GET", "/stalled", nil, nil)
	data, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err == nil || string(data) != "first\n" {
		t.Fatalf("stalled stream should terminate after partial response: %q %v", data, err)
	}
	select {
	case <-stalledDone:
	case <-time.After(time.Second):
		t.Fatal("upstream context was not cancelled on idle timeout")
	}
	for _, method := range []string{"GET", "POST"} {
		var requestBody io.Reader
		if method == http.MethodPost {
			requestBody = strings.NewReader("body")
		}
		res := gatewayRequest(t, gateway, method, "/active", requestBody, nil)
		data, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || string(data) != "0\n1\n2\n3\n4\n5\n6\n7\n" {
			t.Fatalf("active %s stream cut by total deadline: %q %v", method, data, err)
		}
	}
}

func TestChunkedBodyCannotPassLimitAndCancellationPropagates(t *testing.T) {
	var received atomic.Int64
	cancelled := make(chan struct{})
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cancel" {
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("start\n"))
			w.(http.Flusher).Flush()
			<-r.Context().Done()
			close(cancelled)
			return
		}
		n, err := io.Copy(io.Discard, r.Body)
		received.Store(n)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(204)
	}))
	_, gateway := testGateway(t, testDocument(backend.URL), 32)
	reader, writer := io.Pipe()
	go func() { _, _ = writer.Write([]byte(strings.Repeat("x", 128))); _ = writer.Close() }()
	res := gatewayRequest(t, gateway, "POST", "/upload", reader, nil)
	readResponse(t, res)
	if res.StatusCode != 413 {
		t.Fatalf("chunked overflow status=%d, want 413", res.StatusCode)
	}
	if received.Load() > 32 {
		t.Fatalf("oversized body reached upstream: %d", received.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", gateway.URL+"/cancel", nil)
	req.Host = "example.test"
	res, err := gateway.Client().Do(req)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	_, err = bufio.NewReader(res.Body).ReadString('\n')
	if err != nil {
		cancel()
		res.Body.Close()
		t.Fatal(err)
	}
	cancel()
	res.Body.Close()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("client cancellation did not reach upstream")
	}
}

func TestUpgradedConnectionSurvivesConfigReplacement(t *testing.T) {
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, rw, err := http.NewResponseController(w).Hijack()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = rw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n")
		_ = rw.Flush()
		for {
			line, err := rw.ReadString('\n')
			if err != nil {
				return
			}
			_, _ = rw.WriteString(line)
			if rw.Flush() != nil {
				return
			}
		}
	}))
	d := testDocument(backend.URL)
	d.Routes[0].Plugins["circuit-breaker"] = []byte(`{}`)
	e, gateway := testGateway(t, d, 1<<20)
	conn, err := net.DialTimeout("tcp", strings.TrimPrefix(gateway.URL, "http://"), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	_, _ = fmt.Fprint(conn, "GET /websocket HTTP/1.1\r\nHost: example.test\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\nearly-frame\n")
	reader := bufio.NewReader(conn)
	res, err := http.ReadResponse(reader, &http.Request{Method: "GET"})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 101 {
		t.Fatalf("upgrade status %d", res.StatusCode)
	}
	lineBefore, earlyErr := reader.ReadString('\n')
	if earlyErr != nil || lineBefore != "early-frame\n" {
		t.Fatalf("buffered upgrade data lost: %q %v", lineBefore, earlyErr)
	}
	d.Upstreams = map[string]manifest.Upstream{}
	d.Routes[0].Upstream = ""
	d.Routes[0].Plugins = manifest.Plugins{"static-response": []byte(`{"status":200,"body":"replacement"}`)}
	commit(t, e, d)
	_, _ = conn.Write([]byte("after-reload\n"))
	line, err := reader.ReadString('\n')
	if err != nil || line != "after-reload\n" {
		t.Fatalf("upgrade interrupted by reload: %q %v", line, err)
	}
}
