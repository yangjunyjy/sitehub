package publisher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"sitehub/gateway/manifest"
)

func TestPublishUsesAuthenticatedSingleCASDocument(t *testing.T) {
	d := manifest.Document{Version: 1, Routes: []manifest.Route{{ID: "new"}}}
	revision := int64(7)
	puts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/config" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("invalid control request")
			w.WriteHeader(401)
			return
		}
		if r.Method == "PUT" {
			puts++
			if r.Header.Get("If-Match") != `"7"` {
				w.WriteHeader(409)
				return
			}
			var got manifest.Document
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil || got.Routes[0].ID != "new" {
				t.Error("not a complete document")
			}
			revision++
		}
		_ = json.NewEncoder(w).Encode(State{Revision: revision, AppliedRevision: revision, Manifest: d})
	}))
	defer server.Close()
	c := &Client{URL: server.URL, Key: "test-key", HTTP: server.Client()}
	before, err := c.Snapshot(context.Background())
	if err != nil || before.Revision != 7 {
		t.Fatal(before, err)
	}
	after, err := c.Publish(context.Background(), d, 7)
	if err != nil || after.Revision != 8 || puts != 1 {
		t.Fatal(after, err, puts)
	}
	if _, err = c.Publish(context.Background(), d, 6); err == nil {
		t.Fatal("accepted conflicting revision")
	}
	if revision != 8 {
		t.Fatal("conflict changed state")
	}
}
func TestDiffUsesNativeResources(t *testing.T) {
	old := manifest.Document{Routes: []manifest.Route{{ID: "removed"}, {ID: "changed", Host: "old.test"}}}
	next := manifest.Document{Routes: []manifest.Route{{ID: "new"}, {ID: "changed", Host: "new.test"}}}
	got := Diff(next, old)
	want := []string{"+ route/new", "- route/removed", "~ route/changed"}
	if len(got) != len(want) {
		t.Fatal(got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatal(got)
		}
	}
}

func TestDiffIncludesGlobalPluginChanges(t *testing.T) {
	old := manifest.Document{Plugins: manifest.Plugins{"rate-limit": json.RawMessage(`{"rate":10,"burst":10}`)}}
	next := manifest.Document{Plugins: manifest.Plugins{"rate-limit": json.RawMessage(`{"rate":20,"burst":20}`)}}
	if got := Diff(next, old); len(got) != 1 || got[0] != "~ plugins/global" {
		t.Fatalf("global plugin change missing from diff: %v", got)
	}
	if got := Diff(manifest.Document{}, old); len(got) != 1 || got[0] != "- plugins/global" {
		t.Fatalf("global plugin removal missing from diff: %v", got)
	}
}
