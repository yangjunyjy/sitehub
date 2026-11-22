package console

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicListenerStripsConsoleCookieOnly(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add("Cookie", "app=first; "+sessionCookie+"=one")
	req.Header.Add("Cookie", sessionCookie+" =two; preference=second")
	WithoutSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lines := r.Header.Values("Cookie")
		if len(lines) != 2 || lines[0] != "app=first" || lines[1] != "preference=second" {
			t.Fatalf("unexpected forwarded cookies: %v", lines)
		}
	})).ServeHTTP(httptest.NewRecorder(), req)
}
