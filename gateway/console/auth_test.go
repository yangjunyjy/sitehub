package console

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func consoleRequest(c *console, path string, body any, cookie *http.Cookie, origin string) *httptest.ResponseRecorder {
	method := http.MethodGet
	var data []byte
	if body != nil {
		method = http.MethodPost
		data, _ = json.Marshal(body)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(data))
	r.Header.Set("Origin", origin)
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	c.handler().ServeHTTP(w, r)
	return w
}

func loginFixture(t *testing.T, c *console, name string) string {
	t.Helper()
	w := consoleRequest(c, "/api/auth/login", map[string]string{"username": name, "password": "fixture-password"}, nil, c.Origin)
	if w.Code != 204 {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	cookie := w.Result().Cookies()[0]
	if cookie.Name != sessionCookie || !cookie.HttpOnly || cookie.Path != c.CookiePath || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected session cookie attributes: %+v", cookie)
	}
	return cookie.Value
}

func TestIndependentIdentityAndRevocation(t *testing.T) {
	c, _ := testConsole(t)
	c.Origin = "https://console.test"
	token := loginFixture(t, c, "deployer")
	get := func(name, value string) *httptest.ResponseRecorder {
		return consoleRequest(c, "/api/workspace", nil, &http.Cookie{Name: name, Value: value}, c.Origin)
	}
	if w := get("sitehub_admin_session", token); w.Code != 401 {
		t.Fatal("business admin session accepted", w.Code)
	}
	if w := get(sessionCookie, token); w.Code != 200 || !strings.Contains(w.Body.String(), `"publish":true`) || !strings.Contains(w.Body.String(), `"rollback":false`) {
		t.Fatal(w.Code, w.Body.String())
	}
	cookie := &http.Cookie{Name: sessionCookie, Value: token}
	// Guard must reject this role before reading the payload or forwarding to control.
	w := consoleRequest(c, "/api/rollback", map[string]any{}, cookie, c.Origin)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	w = consoleRequest(c, "/api/publish", map[string]any{"revision": 7, "password": "wrong"}, cookie, c.Origin)
	if w.Code != 403 {
		t.Fatal("incorrect confirmation accepted", w.Code)
	}
	w = consoleRequest(c, "/api/auth/logout", map[string]any{}, cookie, "https://other.test")
	if w.Code != 403 || get(sessionCookie, token).Code != 200 {
		t.Fatal("cross-origin logout")
	}
	w = consoleRequest(c, "/api/auth/logout", map[string]any{}, cookie, c.Origin)
	if w.Code != 204 || get(sessionCookie, token).Code != 401 || !w.Result().Cookies()[0].Secure {
		t.Fatal("logout failed")
	}
	token = loginFixture(t, c, "reader")
	if err := SetAccount(c.dir, "reader", "publisher", []byte("fixture-password")); err != nil {
		t.Fatal(err)
	}
	if get(sessionCookie, token).Code != 401 {
		t.Fatal("role change retained previous session")
	}
	token = loginFixture(t, c, "reader")
	if err := RemoveAccount(c.dir, "reader"); err != nil {
		t.Fatal(err)
	}
	if get(sessionCookie, token).Code != 401 {
		t.Fatal("deleted account retained previous session")
	}
}

func TestAccountStorageAndThrottling(t *testing.T) {
	c, _ := testConsole(t)
	path, err := accountFile(c.dir, "operator")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || bytes.Contains(data, []byte("fixture-password")) {
		t.Fatal("credential storage")
	}
	if err := SetAccount(c.dir, "../escape", "admin", []byte("fixture-password")); err == nil {
		t.Fatal("invalid account path accepted")
	}
	for i := 0; i < 10; i++ {
		w := consoleRequest(c, "/api/auth/login", map[string]string{"username": "unknown", "password": strings.Repeat("a", 73)}, nil, c.Origin)
		if w.Code != 403 {
			t.Fatalf("failed login %d: %d", i, w.Code)
		}
	}
	w := consoleRequest(c, "/api/auth/login", map[string]string{"username": "operator", "password": "fixture-password"}, nil, c.Origin)
	if w.Code != 429 {
		t.Fatalf("limit bypassed: %d", w.Code)
	}
}
