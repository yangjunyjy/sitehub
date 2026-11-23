package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"sitehub/gateway/plugin"
)

func TestTrustedProxySuppliesClientAndSchemeToRoutingAndQuota(t *testing.T) {
	backend := testBackend(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(r.Header)
	}))
	d := testDocument(backend.URL)
	d.Routes[0].Host = "example.test:443"
	d.Routes[0].Plugins["limit-count"] = []byte(`{"count":1,"window_seconds":60}`)
	e := New(nil, plugin.DefaultRegistry(), 4096, netip.MustParsePrefix("10.0.0.0/24"))
	t.Cleanup(e.Close)
	commit(t, e, d)
	eventually(t, "upstream ready", func() bool { return countHealthy(e, "app") == 1 })

	request := func(peer, chain, wantIP string, wantStatus int) {
		t.Helper()
		req := httptest.NewRequest("GET", "http://example.test/check", nil)
		req.RemoteAddr = peer + ":12345"
		req.Header.Set("X-Forwarded-For", chain)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Real-IP", "203.0.113.254")
		w := httptest.NewRecorder()
		e.ServeHTTP(w, req)
		if w.Code != wantStatus {
			t.Fatalf("%s via %s got %d, want %d", chain, peer, w.Code, wantStatus)
		}
		if w.Code == 200 {
			var h http.Header
			if err := json.Unmarshal(w.Body.Bytes(), &h); err != nil {
				t.Fatal(err)
			}
			if h.Get("X-Forwarded-For") != wantIP || h.Get("X-Real-IP") != wantIP || h.Get("X-Forwarded-Proto") != "https" {
				t.Fatalf("incorrect forwarding headers %v", h)
			}
		}
	}
	request("10.0.0.2", "192.0.2.22, 198.51.100.1, 10.0.0.3", "198.51.100.1", 200)
	request("10.0.0.2", "198.51.100.2", "198.51.100.2", 200)
	request("10.0.0.2", "198.51.100.1", "", 429)
	request("192.0.2.4", "198.51.100.1", "", 404)
}
