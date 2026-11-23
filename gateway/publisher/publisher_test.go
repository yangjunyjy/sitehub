package publisher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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

func TestReleaseClassifiesControlResponseAndReconcilesLostAcknowledgement(t *testing.T) {
	previous := State{Revision: 7, AppliedRevision: 7, Manifest: manifest.Document{Version: 1}}
	desired := manifest.Document{Version: 1, Routes: []manifest.Route{{ID: "changed"}}}
	for _, tc := range []struct {
		code int
		want string
	}{{400, "rejected"}, {409, "conflict"}, {503, "unconfirmed"}} {
		t.Run(tc.want, func(t *testing.T) {
			var receiptID string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPut:
					receiptID = r.Header.Get("X-Release-ID")
					if !GenerationID.MatchString(receiptID) || r.Header.Get("If-Match") != `"7"` {
						t.Error("missing release ID or revision")
					}
					w.WriteHeader(tc.code)
					_ = json.NewEncoder(w).Encode(map[string]string{"message": "fixture detail"})
				case http.MethodGet:
					if r.URL.Path != "/v1/releases/"+receiptID {
						t.Error("incorrect receipt lookup")
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"revision": 11, "receipt": fmt.Sprintf("7:%s", manifest.Revision(desired))})
				}
			}))
			defer server.Close()
			client := &Client{URL: server.URL, HTTP: server.Client()}
			dir := t.TempDir()
			record, err := ApplyRelease(t.Context(), client, dir, "operator", "", previous, desired)
			var upstream *ControlError
			if !errors.As(err, &upstream) || upstream.Message != "fixture detail" || record.Status != tc.want {
				t.Fatalf("release: %v %v", record, err)
			}
			stored, err := LoadRelease(filepath.Join(dir, record.Generation+".json"))
			if err != nil || stored.Generation != receiptID || stored.Status != tc.want {
				t.Fatalf("persisted release: %v %v", stored, err)
			}
			after := State{Revision: 11, AppliedRevision: 7, Manifest: desired}
			if err := client.Reconcile(t.Context(), &stored, after); err != nil {
				t.Fatal(err)
			}
			if tc.code != 503 {
				if stored.Status != tc.want || stored.Current(after) {
					t.Fatal("rejected release became eligible for rollback")
				}
				return
			}
			if !stored.Current(after) || stored.Status != "committed" {
				t.Fatal("receipt did not establish commit")
			}
			after.AppliedRevision = 11
			if err := client.Reconcile(t.Context(), &stored, after); err != nil || stored.Status != "applied" {
				t.Fatal(stored, err)
			}
			after.Manifest = previous.Manifest
			if stored.Current(after) {
				t.Fatal("unrelated content at same revision accepted for rollback")
			}
		})
	}
}

func TestUnconfirmedReleaseWithoutMatchingReceipt(t *testing.T) {
	for _, code := range []int{200, 404} {
		r := Release{ReleaseSummary: ReleaseSummary{Status: "pending", Generation: "000000000000000000000000", PreviousRevision: 4}, Desired: manifest.Document{Version: 1}}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(code)
			_ = json.NewEncoder(w).Encode(map[string]any{"revision": 5, "receipt": "4:different-manifest"})
		}))
		c := &Client{URL: server.URL, HTTP: server.Client()}
		err := c.Reconcile(t.Context(), &r, State{Revision: 5, AppliedRevision: 5, Manifest: r.Desired})
		server.Close()
		if r.Status != "pending" || (err != nil) != (code == 200) {
			t.Fatalf("code %d: %v %v", code, r, err)
		}
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

func TestDiffIgnoresPluginObjectKeyOrder(t *testing.T) {
	doc := func(raw string) manifest.Document {
		return manifest.Document{Plugins: manifest.Plugins{"static-response": json.RawMessage(raw)}}
	}
	before := doc(`{"status":200,"body":"ok"}`)
	if changes := Diff(doc(`{"body":"ok","status":200}`), before); len(changes) != 0 {
		t.Fatalf("reordered object reported as changed: %v", changes)
	}
	if changes := Diff(doc(`{"body":"ok","status":201}`), before); len(changes) != 1 {
		t.Fatalf("changed status missing: %v", changes)
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
