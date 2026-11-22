package engine

import (
	"encoding/json"
	"net/http"
	"testing"

	"sitehub/gateway/manifest"
)

func TestHTTPForwardingByAuthorityAndPath(t *testing.T) {
	upstream := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	d := testDocument(upstream.URL)
	d.Routes = nil
	respond := func(id, host, path string, priority int) {
		config, _ := json.Marshal(map[string]any{"status": 200, "body": id})
		d.Routes = append(d.Routes, manifest.Route{ID: id, Host: host, Paths: []string{path}, Priority: priority, Plugins: manifest.Plugins{"static-response": config}})
	}
	respond("fallback", "", "/*", 0)
	respond("any-port", "*:8443", "/*", 0)
	respond("ipv4", "127.0.0.1", "/*", 0)
	respond("loopback-port", "127.0.0.1:8443", "/*", 0)
	respond("ipv6", "[::1]", "/*", 0)
	respond("ipv6-port", "[::1]:8443", "/*", 0)
	respond("default-port", "example.test:80", "/*", 0)
	respond("port-health", "*", "/health", 10)
	d.Routes = append(d.Routes, manifest.Route{ID: "forward", Host: "127.0.0.1", Methods: []string{"GET"}, Paths: []string{"/api/*"}, Upstream: "app"})
	_, gateway := testGateway(t, d, 1<<20)
	for _, tc := range []struct{ method, authority, path, want string }{
		{"GET", "127.0.0.1:9999", "/api/info", "/api/info"},
		{"POST", "127.0.0.1:9999", "/api/info", "ipv4"},
		{"GET", "127.0.0.1:8443", "/api/info", "loopback-port"},
		{"GET", "[0:0:0:0:0:0:0:1]:8443", "/", "ipv6-port"},
		{"GET", "[::1]", "/", "ipv6"},
		{"GET", "other.test:8443", "/", "any-port"},
		{"GET", "Example.Test.", "/", "default-port"},
		{"GET", "127.0.0.1:8443", "/health", "port-health"},
		{"GET", "other.test", "/", "fallback"},
	} {
		req, err := http.NewRequest(tc.method, gateway.URL+tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = tc.authority
		resp, err := gateway.Client().Do(req)
		if err != nil {
			t.Fatal(err)
		}
		if got := readResponse(t, resp); got != tc.want {
			t.Errorf("%s %s%s: got %q, want %q", tc.method, tc.authority, tc.path, got, tc.want)
		}
	}
}
