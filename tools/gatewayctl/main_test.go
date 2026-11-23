package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sitehub/gateway/manifest"
	"sitehub/gateway/publisher"
)

func TestRollbackReconcilesUnknownReleaseAndChecksCurrentContent(t *testing.T) {
	doc := func(id string) manifest.Document {
		return manifest.Document{Version: 1, Routes: []manifest.Route{{ID: id, Host: "*", Paths: []string{"/*"}, Plugins: manifest.Plugins{"static-response": json.RawMessage(`{"status":204}`)}}}}
	}
	id := "111111111111111111111111"
	desired, original := doc("published"), doc("previous")
	state := publisher.State{Revision: 12, AppliedRevision: 12, Manifest: desired}
	puts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/releases/"+id {
			_ = json.NewEncoder(w).Encode(map[string]any{"revision": 12, "receipt": "7:" + manifest.Revision(desired)})
			return
		}
		if r.Method == http.MethodPut {
			puts++
			if r.Header.Get("If-Match") != `"12"` || !publisher.GenerationID.MatchString(r.Header.Get("X-Release-ID")) {
				t.Error("rollback missing CAS or new release ID")
			}
			var received manifest.Document
			if err := json.NewDecoder(r.Body).Decode(&received); err != nil || manifest.Revision(received) != manifest.Revision(original) {
				t.Error("wrong rollback target")
			}
			state.Manifest = received
			state.Revision = 13
			state.AppliedRevision = 13
		}
		_ = json.NewEncoder(w).Encode(state)
	}))
	defer server.Close()
	dir := t.TempDir()
	t.Setenv("GATEWAYCTL_ADMINURL", server.URL)
	t.Setenv("GATEWAYCTL_ADMINKEY", strings.Repeat("a", 32))
	t.Setenv("GATEWAYCTL_STATEDIR", dir)
	config, release := filepath.Join(dir, "config.yaml"), filepath.Join(dir, id+".json")
	if err := os.WriteFile(config, []byte("AdminURL: local\n"), 0600); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(publisher.Release{ReleaseSummary: publisher.ReleaseSummary{Status: "unconfirmed", Generation: id, PreviousRevision: 7}, Previous: original, Desired: desired})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(release, data, 0600); err != nil {
		t.Fatal(err)
	}
	run := func(action string) (string, error) {
		cmd := newCommand()
		var output bytes.Buffer
		cmd.SetOut(&output)
		cmd.SetArgs([]string{action, "-f", config, "--release", release})
		err := cmd.ExecuteContext(t.Context())
		return output.String(), err
	}
	if out, err := run("reconcile"); err != nil || !strings.Contains(out, fmt.Sprintf("release %s status applied revision 12", id)) {
		t.Fatal(out, err)
	}
	state.Manifest = doc("wrong")
	if _, err := run("rollback"); err == nil || puts != 0 {
		t.Fatal("rolled back unrelated content")
	}
	state.Manifest = desired
	if out, err := run("rollback"); err != nil || puts != 1 {
		t.Fatal(out, err)
	}
	if _, err := run("rollback"); err == nil || puts != 1 {
		t.Fatal("rollback overwrote later revision")
	}
}
