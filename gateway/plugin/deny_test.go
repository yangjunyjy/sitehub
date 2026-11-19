package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDenyPathsMatchesExactAndSegmentPrefix(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := build(t, `{"request-id":{},"deny-paths":{"paths":["/healthz","/internal","/internal/*"]}}`, next)
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/healthz", 404},
		{"HEAD", "/healthz", 404},
		{"POST", "/internal", 404},
		{"GET", "/internal/", 404},
		{"GET", "/internal/config", 404},
		{"GET", "/internal/nested/config", 404},
		{"GET", "/internalized", 204},
		{"GET", "/healthz/live", 204},
		{"GET", "/other/internal", 204},
		{"GET", "/public?redirect=/internal/config", 204},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != test.status {
			t.Errorf("%s %s: got %d, want %d", test.method, test.path, response.Code, test.status)
		}
		if test.method == http.MethodHead && response.Body.Len() != 0 {
			t.Fatal("denied HEAD response contains a body")
		}
		if response.Header().Get("X-Request-ID") == "" {
			t.Fatal("deny-paths ran before request-id")
		}
	}
}

func TestDenyPathsExplicitStatusAndWildcardSemantics(t *testing.T) {
	for _, test := range []struct {
		config, path string
		status       int
	}{
		{`{"deny-paths":{"paths":["/private/*"],"status":403}}`, "/private/config", 403},
		{`{"deny-paths":{"paths":["/private/*"],"status":403}}`, "/private", 204},
		{`{"deny-paths":{"paths":["/private/*"],"status":403}}`, "/private-data", 204},
		{`{"deny-paths":{"paths":["/*"],"status":503}}`, "/anything", 503},
		{`{"deny-paths":{"paths":["/*"],"status":503}}`, "/", 503},
	} {
		handler := build(t, test.config, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
		if response.Code != test.status {
			t.Errorf("%s %s: got %d, want %d", test.config, test.path, response.Code, test.status)
		}
	}
}

func TestDenyPathsRejectsAmbiguousConfiguration(t *testing.T) {
	for _, raw := range []string{
		`null`, `{}`, `{"paths":null}`, `{"paths":[]}`, `{"paths":["/private"],"unknown":true}`,
		`{"paths":["/private"],"status":0}`, `{"paths":["/private"],"status":399}`, `{"paths":["/private"],"status":600}`,
		`{"paths":["private"]}`, `{"paths":["/private","/private"]}`, `{"paths":["/private*"]}`, `{"paths":["/private/*/other"]}`,
		`{"paths":["/private/../public"]}`, `{"paths":["/private/./public"]}`, `{"paths":["/private//public"]}`,
		`{"paths":["/private?query"]}`, `{"paths":["/private#fragment"]}`, `{"paths":["/private%2Fpublic"]}`,
		`{"paths":["/private\\public"]}`, `{"paths":["/private\r\npublic"]}`, `{"paths":["/private\tpublic"]}`,
	} {
		if _, err := denyPaths(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted invalid deny-paths configuration %s", raw)
		}
	}
}
