package control

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"sitehub/gateway/manifest"
)

func TestFileReloadRejectsInvalidCandidateAndKeepsFilesAuthoritative(t *testing.T) {
	store := &MemoryStore{}
	if _, err := Seed(context.Background(), store, staticDocument("old")); err != nil {
		t.Fatal(err)
	}
	c := newTestController(t, store)
	desired := staticDocument("new")
	var loadErr error
	c.Load = func() (manifest.Document, error) { return desired, loadErr }
	if response := adminRequest(c, http.MethodPut, "/v1/config", manifest.Bytes(desired), `"1"`); response.Code != 405 {
		t.Fatalf("file mode accepted remote publication: %d", response.Code)
	}
	assertLive(t, c, "old")
	loadErr = errors.New("invalid included YAML")
	if response := adminRequest(c, http.MethodPost, "/v1/reload", nil, ""); response.Code != 400 {
		t.Fatalf("invalid file was accepted: %d", response.Code)
	}
	assertLive(t, c, "old")
	loadErr = nil
	desired.Routes[1].Plugins["static-response"] = json.RawMessage(`{"status":999}`)
	if response := adminRequest(c, http.MethodPost, "/v1/reload", nil, ""); response.Code != 400 {
		t.Fatalf("invalid plugin was accepted: %d", response.Code)
	}
	assertLive(t, c, "old")
	before, _ := store.Read(context.Background())
	if before.Revision != 1 {
		t.Fatal("failed reload changed stored revision")
	}
	desired = staticDocument("new")
	if response := adminRequest(c, http.MethodPost, "/v1/reload", nil, ""); response.Code != 200 {
		t.Fatalf("valid reload failed: %d %s", response.Code, response.Body.String())
	}
	assertLive(t, c, "new")
	if response := adminRequest(c, http.MethodPost, "/v1/reload", nil, ""); response.Code != 200 {
		t.Fatal("unchanged reload failed")
	}
	after, _ := store.Read(context.Background())
	if after.Revision != 2 {
		t.Fatal("unchanged reload advanced revision")
	}
}

func TestFileReloadCannotBypassControlAuthenticationOrEtcdPublication(t *testing.T) {
	store := &memoryStore{record: Record{Revision: 1, Data: manifest.Bytes(staticDocument("old"))}}
	c := newTestController(t, store)
	if response := adminRequest(c, http.MethodPost, "/v1/reload", nil, ""); response.Code != 409 {
		t.Fatal("etcd mode accepted file reload")
	}
	c.Load = func() (manifest.Document, error) {
		t.Fatal("unauthorized request read files")
		return manifest.Document{}, nil
	}
	for _, remote := range []string{"172.31.0.20:50000", "203.0.113.1:50000"} {
		req := httptest.NewRequest(http.MethodPost, "/v1/reload", nil)
		req.RemoteAddr = remote
		response := httptest.NewRecorder()
		c.Handler().ServeHTTP(response, req)
		if response.Code != 401 && response.Code != 403 {
			t.Fatal("reload bypassed control access checks")
		}
	}
}
