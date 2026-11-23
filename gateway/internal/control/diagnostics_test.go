package control

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"sitehub/gateway/diagnostics"
	"sitehub/gateway/manifest"
)

func TestInspectAndForwardUseTheSameAppliedRules(t *testing.T) {
	doc := staticDocument("ok")
	doc.Routes = nil
	add := func(id, pattern string, priority int) {
		settings, _ := json.Marshal(map[string]any{"body": id, "status": 200})
		doc.Routes = append(doc.Routes, manifest.Route{ID: id, Host: "blog.sitehub.local:8082", Methods: []string{"GET"}, Paths: []string{pattern}, Priority: priority, Plugins: manifest.Plugins{"static-response": settings}})
	}
	add("fallback", "/*", 0)
	add("home", "/", 0)
	add("news", "/news/*", 0)
	add("priority", "/staff/*", 1)
	add("ignored", "/staff/index", 0)
	c := newTestController(t, &memoryStore{record: Record{Revision: 7, Data: manifest.Bytes(doc)}})
	for _, tc := range []struct {
		method, host, path, rule, pattern, miss string
		code                                    int
	}{
		{"GET", "blog.sitehub.local:8082", "/?preview=1", "home", "/", "", 200},
		{"GET", "blog.sitehub.local:8082", "/a", "fallback", "/*", "", 200},
		{"GET", "blog.sitehub.local:8082", "/news/today", "news", "/news/*", "", 200},
		{"GET", "blog.sitehub.local:8082", "/staff/index", "priority", "/staff/*", "", 200},
		{"GET", "blog.ops-gofer.cn", "/", "", "", "Host", 404},
		{"POST", "blog.sitehub.local:8082", "/", "", "", "方法", 404},
		{"GET", "blog.sitehub.local:8082", "/news%2ftoday", "", "", "", 400},
	} {
		t.Run(tc.method+" "+tc.host+tc.path, func(t *testing.T) {
			payload, _ := json.Marshal(diagnostics.MatchRequest{Scheme: "http", Host: tc.host, Path: tc.path, Method: tc.method})
			inspected := adminRequest(c, "POST", "/v1/match", payload, "")
			live := httptest.NewRecorder()
			c.Engine.ServeHTTP(live, httptest.NewRequest(tc.method, "http://"+tc.host+tc.path, nil))
			if live.Code != tc.code || tc.rule != "" && live.Body.String() != tc.rule {
				t.Fatalf("live = %d %s, want %d %s", live.Code, live.Body.String(), tc.code, tc.rule)
			}
			if tc.code == 400 {
				if inspected.Code != 400 {
					t.Fatalf("unsafe path accepted: %s", inspected.Body.String())
				}
				return
			}
			var got diagnostics.MatchResult
			if err := json.Unmarshal(inspected.Body.Bytes(), &got); err != nil || inspected.Code != 200 || got.Revision != 7 || got.Matched != (tc.rule != "") || got.RouteID != tc.rule || got.Pattern != tc.pattern || !strings.Contains(got.Reason, tc.miss) {
				t.Fatalf("inspection = %d %+v, %v", inspected.Code, got, err)
			}
			if got.Matched && got.Terminal != "static-response" {
				t.Fatalf("missing response plugin: %+v", got)
			}
		})
	}
	if warnings := manifest.PathWarnings(doc.Routes); len(warnings) != 1 || !strings.Contains(warnings[0], "home") || !strings.Contains(warnings[0], "fallback") {
		t.Fatalf("migration warnings = %v", warnings)
	}
}

func TestRuntimeReflectsAppliedSocketsAndNodesAfterRejectedWatch(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	defer upstream.Close()
	listen, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listen.LocalAddr().String()
	listen.Close()
	d := staticDocument("live")
	d.Streams = []manifest.Stream{{ID: "dns", Protocol: "udp", Listen: address, Target: "127.0.0.1:53"}}
	d.Upstreams = map[string]manifest.Upstream{"app": {Nodes: []manifest.Node{{URL: upstream.URL, Weight: 2}}, ConnectTimeoutMS: 1000, ResponseHeaderTimeoutMS: 1000, IOTimeoutMS: 1000, HealthPath: "/healthz", HealthIntervalMS: 1000, HealthTimeoutMS: 500, HealthFailures: 1}}
	store := &memoryStore{record: Record{Revision: 11, Data: manifest.Bytes(d)}}
	c := newTestController(t, store)
	c.HTTPListen = "127.0.0.1:9080"
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	next := staticDocument("desired")
	next.Streams = []manifest.Stream{{ID: "busy", Protocol: "tcp", Listen: occupied.Addr().String(), Target: "127.0.0.1:80"}}
	store.record = Record{Revision: 12, Data: manifest.Bytes(next)}
	if err := c.Apply(store.record); err == nil {
		t.Fatal("occupied port was applied")
	}
	var state State
	for until := time.Now().Add(4 * time.Second); ; time.Sleep(10 * time.Millisecond) {
		res := adminRequest(c, "GET", "/v1/config", nil, "")
		if err := json.Unmarshal(res.Body.Bytes(), &state); err != nil || res.Code != 200 || state.Runtime == nil {
			t.Fatalf("config = %d %s", res.Code, res.Body.String())
		}
		nodes := state.Runtime.Upstreams["app"]
		if len(nodes) == 1 && nodes[0].URL == upstream.URL && nodes[0].Healthy && nodes[0].Weight == 2 {
			break
		}
		if time.Now().After(until) {
			t.Fatalf("backend health = %+v", nodes)
		}
	}
	if state.Revision != 12 || state.AppliedRevision != 11 || state.LastError == "" || state.AppliedManifest == nil || state.AppliedManifest.Streams[0].ID != "dns" || state.Manifest.Streams[0].ID != "busy" {
		t.Fatalf("inconsistent applied vs desired state: %+v", state)
	}
	r := state.Runtime
	if r.HTTPListen != c.HTTPListen || len(r.Streams) != 1 || r.Streams[0].Address != address || r.Streams[0].Protocol != "udp" {
		t.Fatalf("runtime bindings = %+v", r)
	}
	assertLive(t, c, "live")
	store.record = Record{}
	if w := adminRequest(c, "GET", "/v1/config", nil, ""); w.Code != http.StatusServiceUnavailable {
		t.Fatalf("publisher received a writable base after deletion: %s", w.Body.String())
	}
	w := adminRequest(c, "GET", "/v1/config?local_fallback=1", nil, "")
	var fallback State
	if err := json.Unmarshal(w.Body.Bytes(), &fallback); err != nil || w.Code != http.StatusOK || !fallback.DesiredUnavailable || fallback.Revision != 0 || fallback.AppliedRevision != 11 || fallback.AppliedManifest == nil || fallback.Runtime == nil || fallback.LastError == "" || w.Header().Get("ETag") != "" {
		t.Fatalf("read-only workspace failed: %d %s", w.Code, w.Body.String())
	}
}
