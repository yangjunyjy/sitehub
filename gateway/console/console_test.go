package console

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
	"sitehub/gateway/publisher"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func testConsole(t *testing.T) (*console, *publisher.State) {
	t.Helper()
	state := &publisher.State{Revision: 7, AppliedRevision: 7, Manifest: manifest.Document{
		Version: 1, Routes: []manifest.Route{{ID: "home", Host: "app.test", Paths: []string{"/*"}, Plugins: manifest.Plugins{"static-response": json.RawMessage(`{"status":200,"body":"ok"}`)}}},
	}}
	var mu sync.Mutex
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.URL.Path != "/v1/config" || r.Header.Get("Authorization") != "Bearer fixture-control-key" {
			t.Error("invalid gateway request")
			w.WriteHeader(401)
			return
		}
		if r.Method == "PUT" {
			if r.Header.Get("If-Match") != strconv.Quote(strconv.FormatInt(state.Revision, 10)) {
				w.WriteHeader(409)
				return
			}
			var next manifest.Document
			if err := json.NewDecoder(r.Body).Decode(&next); err != nil {
				t.Error(err)
				return
			}
			state.Manifest = next
			state.Revision++
			state.AppliedRevision = state.Revision
		}
		_ = json.NewEncoder(w).Encode(state)
	}))
	t.Cleanup(gateway.Close)
	dir := t.TempDir()
	auth, err := newIdentityStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	for name, role := range map[string]string{"operator": "admin", "reader": "viewer", "deployer": "publisher"} {
		if err := SetAccount(dir, name, role, []byte("fixture-password")); err != nil {
			t.Fatal(err)
		}
	}
	return &console{Config: Config{Origin: "http://console.test", CookiePath: "/gateway-console/"}, identity: auth, plugins: plugin.DefaultRegistry(), client: &publisher.Client{URL: gateway.URL, Key: "fixture-control-key", HTTP: gateway.Client()}, dir: dir}, state
}

func TestConsoleReleaseLifecycle(t *testing.T) {
	c, state := testConsole(t)
	handler := c.handler()
	tokens := map[string]string{}
	for _, name := range []string{"operator", "reader"} {
		tokens[name] = loginFixture(t, c, name)
	}
	request := func(path, token string, body any, want int) map[string]any {
		t.Helper()
		method := "GET"
		if body != nil {
			method = "POST"
		}
		data, _ := json.Marshal(body)
		r := httptest.NewRequest(method, "/api/"+path, bytes.NewReader(data))
		if token != "" {
			r.AddCookie(&http.Cookie{Name: sessionCookie, Value: tokens[token]})
		}
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", c.Origin)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	request("workspace", "", nil, 401)
	if view := request("workspace", "reader", nil, 200); view["publish"] != false {
		t.Fatal(view)
	}
	original := manifest.Revision(state.Manifest)
	desired, _ := manifest.Decode(manifest.Bytes(state.Manifest))
	desired.Routes[0].Priority = 3
	payload := map[string]any{"password": "fixture-password", "revision": 7, "manifest": desired}
	request("publish", "reader", payload, 403)
	request("preview", "operator", map[string]any{"revision": 7, "manifest": map[string]any{"version": 9}}, 400)
	r := httptest.NewRequest("POST", "/api/publish", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("missing origin accepted")
	}
	request("preview", "operator", payload, 200)
	if state.Revision != 7 {
		t.Fatal("preview changed revision")
	}
	result := request("publish", "operator", payload, 200)
	if state.Revision != 8 {
		t.Fatal("publish failed")
	}
	request("publish", "operator", payload, 409)
	rollback := map[string]any{"password": "fixture-password", "generation": result["generation"], "revision": 8}
	request("rollback", "operator", rollback, 200)
	if state.Revision != 9 || manifest.Revision(state.Manifest) != original {
		t.Fatal("rollback did not restore original configuration")
	}
	rollback["revision"] = 9
	request("rollback", "operator", rollback, 409)
	history, err := releaseHistory(c.dir)
	if err != nil || len(history) != 2 || history[0].RollbackOf != result["generation"] || history[1].Actor != "operator" {
		t.Fatal(history, err)
	}
	state.FileManaged = true
	if view := request("workspace", "operator", nil, 200); view["publish"] != false || view["rollback"] != false {
		t.Fatal("file mode offered writes")
	}
	payload["revision"] = 9
	request("publish", "operator", payload, 409)
}

func TestConsoleBrowser(t *testing.T) {
	if os.Getenv("GATEWAY_CONSOLE_BROWSER_TEST") != "1" {
		t.Skip("set GATEWAY_CONSOLE_BROWSER_TEST=1 to run Playwright")
	}
	c, _ := testConsole(t)
	mux := http.NewServeMux()
	mux.Handle("/gateway-console/", http.StripPrefix("/gateway-console", c.handler()))
	server := httptest.NewUnstartedServer(mux)
	c.Origin = "http://" + server.Listener.Addr().String()
	server.Start()
	defer server.Close()
	cmd := exec.Command("node", "../../web/tests/gateway-console-browser.mjs")
	cmd.Env = append(os.Environ(), "GATEWAY_CONSOLE_TEST_URL="+server.URL+"/gateway-console/")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: %s", err, strings.TrimSpace(string(output)))
	}
}
