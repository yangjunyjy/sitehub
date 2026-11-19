package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"testing"
)

func TestProxyRewritePreservesEscapingQueryAndOriginalRequest(t *testing.T) {
	for _, test := range []struct {
		name, prefix, target, wantPath, wantEscaped string
	}{
		{"encoded prefix and suffix", "/api", "http://public.example/%61pi/a%20b/%e4%b8%ad?sig=a%2Bb&q=c+d", "/a b/中", "/a%20b/%e4%b8%ad"},
		{"unicode prefix", "/接口", "http://public.example/%E6%8E%A5%E5%8F%A3/a%20b?q=%2F", "/a b", "/a%20b"},
		{"exact prefix becomes root", "/api", "http://public.example/api?sig=%3D", "/", "/"},
		{"segment boundary", "/api", "http://public.example/apiv2/data?q=%2F", "/apiv2/data", "/apiv2/data"},
		{"trailing slash", "/api", "http://public.example/api/?q=%2F", "/", "/"},
		{"unmatched path", "/api", "http://public.example/other/api?q=%2F", "/other/api", "/other/api"},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw, _ := json.Marshal(map[string]any{"strip_prefix": test.prefix, "host": "backend.internal:8081", "headers": map[string]string{"X-Upstream-Region": "east"}})
			middleware, err := proxyRewrite(raw)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodGet, test.target, nil)
			request.Header.Set("X-Upstream-Region", "original")
			request.Header.Set("X-Other", "keep")
			originalURL, originalHost, originalURI := *request.URL, request.Host, request.RequestURI
			originalHeader := request.Header.Clone()
			handler := middleware(http.HandlerFunc(func(_ http.ResponseWriter, received *http.Request) {
				if received == request || received.URL == request.URL {
					t.Fatal("rewrite did not clone request and URL")
				}
				if received.URL.Path != test.wantPath || received.URL.EscapedPath() != test.wantEscaped || received.URL.RawQuery != originalURL.RawQuery {
					t.Fatalf("unexpected rewritten URL: %s", received.URL.String())
				}
				if received.Host != "backend.internal:8081" || received.URL.Host != originalURL.Host {
					t.Fatalf("host rewrite changed upstream destination: host=%q URL.Host=%q", received.Host, received.URL.Host)
				}
				if received.Header.Get("X-Upstream-Region") != "east" || received.Header.Get("X-Other") != "keep" {
					t.Fatalf("unexpected rewritten headers: %v", received.Header)
				}
			}))
			handler.ServeHTTP(httptest.NewRecorder(), request)
			if *request.URL != originalURL || request.Host != originalHost || request.RequestURI != originalURI || !reflect.DeepEqual(request.Header, originalHeader) {
				t.Fatal("rewrite modified the original request")
			}
		})
	}
}

func TestProxyRewriteForwardingThroughReverseProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "service.internal" || r.RequestURI != "/files/a%2Bb?signature=a%2Bb&q=c+d" || r.Header.Get("X-Region") != "east" {
			t.Errorf("upstream received unexpected request: host=%q uri=%q headers=%v", r.Host, r.RequestURI, r.Header)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer backend.Close()
	target, _ := url.Parse(backend.URL)
	proxy := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
		r.Out.URL.Scheme, r.Out.URL.Host = target.Scheme, target.Host
		r.Out.Host = r.In.Host
	}}
	handler := build(t, `{"proxy-rewrite":{"strip_prefix":"/api","host":"service.internal","headers":{"X-Region":"east"}}}`, proxy)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://public.example/api/files/a%2Bb?signature=a%2Bb&q=c+d", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("forwarding failed: %d %s", response.Code, response.Body.String())
	}
}

func TestProxyRewriteRejectsInvalidConfiguration(t *testing.T) {
	for _, raw := range []string{
		`null`, `{"prefix":"/api"}`, `{"strip_prefix":"api"}`, `{"strip_prefix":"/api/"}`, `{"strip_prefix":"/api/../other"}`,
		`{"strip_prefix":"/api?query"}`, `{"strip_prefix":"/api%2F"}`, `{"strip_prefix":"/api\\other"}`, `{"strip_prefix":"/api\n"}`,
		`{"host":"https://backend.internal"}`, `{"host":"backend.internal/path"}`, `{"host":"user@backend.internal"}`,
		`{"host":"backend.internal:0"}`, `{"host":"backend.internal:65536"}`, `{"host":"backend.internal:"}`, `{"host":"backend.internal:abc"}`,
		`{"host":"backend.internal\r\nX-Injected: true"}`, `{"host":"::1"}`, `{"host":"[bad]"}`, `{"host":"backend..internal"}`,
		`{"headers":{"X-Test\n":"value"}}`, `{"headers":{"X-Test":"value\r\nX-Injected:true"}}`, `{"headers":{"X-Test":"one","x-test":"two"}}`,
	} {
		if _, err := proxyRewrite(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted invalid configuration %s", raw)
		}
	}
	for _, name := range []string{"Host", "Content-Length", "Connection", "Transfer-Encoding", "Upgrade", "TE", "Trailer", "Keep-Alive", "Proxy-Connection", "Proxy-Authorization", "Authorization", "Cookie", "X-Service-Key", "X-Actor-ID", "X-User-ID", "X-Role", "X-Roles", "Forwarded", "X-Forwarded-For", "X-Forwarded-Custom", "X-Real-IP", "X-Request-ID"} {
		raw, _ := json.Marshal(map[string]any{"headers": map[string]string{name: "value"}})
		if _, err := proxyRewrite(raw); err == nil {
			t.Errorf("accepted protected header %s", name)
		}
	}
	for _, host := range []string{"backend.internal", "service_name:8080", "127.0.0.1:9080", "[::1]", "[2001:db8::1]:8443"} {
		raw, _ := json.Marshal(map[string]string{"host": host})
		if _, err := proxyRewrite(raw); err != nil {
			t.Errorf("rejected valid host %q: %v", host, err)
		}
	}
}
