package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrowserSessionCustomCookieAndOriginPaths(t *testing.T) {
	middleware, err := browserSession(json.RawMessage(`{
		"cookie_name":"project_session",
		"origins":["https://portal.example"],
		"origin_check_paths":["/session/sign-in","/session/sign-up"]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	var receivedAuthorization, receivedCookie string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthorization, receivedCookie = r.Header.Get("Authorization"), r.Header.Get("Cookie")
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, test := range []struct {
		name, method, path, cookie, authorization, origin, wantAuthorization string
		wantStatus                                                           int
	}{
		{"custom cookie", "GET", "/orders", "project_session=custom-token; sitehub_admin_session=other", "", "", "Bearer custom-token", http.StatusNoContent},
		{"different cookie", "GET", "/orders", "sitehub_admin_session=other", "", "", "", http.StatusNoContent},
		{"explicit authorization", "POST", "/orders", "project_session=custom-token", "Bearer explicit", "", "Bearer explicit", http.StatusNoContent},
		{"cookie write needs origin", "POST", "/orders", "project_session=custom-token", "", "", "", http.StatusForbidden},
		{"cookie write approved", "POST", "/orders", "project_session=custom-token", "", "https://portal.example", "Bearer custom-token", http.StatusNoContent},
		{"custom login foreign origin", "POST", "/session/sign-in", "", "Bearer explicit", "https://foreign.example", "", http.StatusForbidden},
		{"custom registration foreign origin", "POST", "/session/sign-up", "", "", "https://foreign.example", "", http.StatusForbidden},
		{"custom login approved", "POST", "/session/sign-in", "", "", "https://portal.example", "", http.StatusNoContent},
		{"CLI custom login", "POST", "/session/sign-in", "", "", "", "", http.StatusNoContent},
		{"replaced default login paths", "POST", "/v1/rbac/auth/login", "", "", "https://foreign.example", "", http.StatusNoContent},
		{"duplicate custom cookie", "GET", "/orders", "project_session=one; project_session=two", "", "", "", http.StatusBadRequest},
		{"invalid custom cookie value", "GET", "/orders", "project_session=bad token", "", "", "", http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			receivedAuthorization, receivedCookie = "", ""
			req := httptest.NewRequest(test.method, "http://gateway.test"+test.path, nil)
			req.Header.Set("Cookie", test.cookie)
			req.Header.Set("Authorization", test.authorization)
			req.Header.Set("Origin", test.origin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if response.Code == http.StatusNoContent && (receivedAuthorization != test.wantAuthorization || receivedCookie != "") {
				t.Fatalf("upstream credentials = %q, cookie = %q", receivedAuthorization, receivedCookie)
			}
		})
	}
}

func TestBrowserSessionExplicitEmptyOriginPathsStillProtectsCookieWrites(t *testing.T) {
	middleware, err := browserSession(json.RawMessage(`{
		"audience":"admin",
		"cookie_name":"operator_session",
		"origins":["https://admin.example"],
		"origin_check_paths":[]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, test := range []struct {
		cookie string
		want   int
	}{{"", http.StatusNoContent}, {"operator_session=token", http.StatusForbidden}} {
		req := httptest.NewRequest(http.MethodPost, "http://gateway.test/v1/rbac/auth/login", nil)
		req.Header.Set("Cookie", test.cookie)
		req.Header.Set("Origin", "https://foreign.example")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != test.want {
			t.Fatalf("cookie %q: status = %d, want %d", test.cookie, response.Code, test.want)
		}
	}
}

func TestBrowserSessionRejectsInvalidCustomConfiguration(t *testing.T) {
	for _, test := range []struct {
		config, wantError string
	}{
		{`{"origins":["https://portal.example"]}`, "cookie_name or audience"},
		{`{"audience":"other","cookie_name":"session","origins":["https://portal.example"]}`, "audience"},
		{`{"cookie_name":"bad cookie","origins":["https://portal.example"]}`, "cookie_name"},
		{`{"cookie_name":"bad;cookie","origins":["https://portal.example"]}`, "cookie_name"},
		{`{"cookie_name":"bad=cookie","origins":["https://portal.example"]}`, "cookie_name"},
		{`{"cookie_name":"会话","origins":["https://portal.example"]}`, "cookie_name"},
		{`{"cookie_name":"session","origins":["https://portal.example"],"cookie_names":[]}`, "unknown field"},
	} {
		if _, err := browserSession(json.RawMessage(test.config)); err == nil || !strings.Contains(err.Error(), test.wantError) {
			t.Errorf("configuration %s: error = %v, want %q", test.config, err, test.wantError)
		}
	}
	for _, value := range []string{"", "session/login", "//session/login", "/session/../login", "/session//login", "/login?next=/", "/login#form", "/login/%2f", "/login/*", "/log in", "/login\\path", "/login\n", "/login\x00", "/login\x7f"} {
		raw, err := json.Marshal(browserSessionConfig{CookieName: "session", Origins: []string{"https://portal.example"}, OriginCheckPaths: []string{value}})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := browserSession(raw); err == nil || !strings.Contains(err.Error(), "origin_check_paths") {
			t.Errorf("origin_check_paths %q: error = %v", value, err)
		}
	}
	if _, err := browserSession(json.RawMessage(`{"cookie_name":"session","origins":["https://portal.example"],"origin_check_paths":["/login","/login"]}`)); err == nil {
		t.Error("duplicate origin_check_paths accepted")
	}
}
