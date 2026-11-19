package plugin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCustomTerminalPluginNeedsOnlyRegistration(t *testing.T) {
	r := NewRegistry()
	calls := 0
	err := r.RegisterTerminal("maintenance", 100, func(raw json.RawMessage) (Middleware, error) {
		calls++
		var config struct {
			Status int `json:"status"`
		}
		if err := decode(raw, &config); err != nil {
			return nil, err
		}
		return func(http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(config.Status) })
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	configs := map[string]json.RawMessage{"maintenance": []byte(`{"status":503}`)}
	terminal, err := r.Terminal(configs)
	if err != nil || !terminal || calls != 0 {
		t.Fatalf("terminal lookup must use metadata only: terminal=%v calls=%d err=%v", terminal, calls, err)
	}
	middleware, err := r.Build(configs)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	middleware(nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("custom terminal did not respond: %d", response.Code)
	}
	if err := r.Register("maintenance", 200, requestID); err == nil {
		t.Fatal("terminal registration could be overwritten")
	}
	if err := r.RegisterTerminal("other", 50, requestID); err != nil {
		t.Fatal(err)
	}
	configs["other"] = []byte(`{}`)
	if _, err := r.Terminal(configs); err == nil {
		t.Fatal("multiple terminal plugins accepted")
	}
	if _, err := r.Build(configs); err == nil {
		t.Fatal("multiple terminal plugins built")
	}
	configs["other"] = []byte(`null`)
	if terminal, err := r.Terminal(configs); err != nil || !terminal {
		t.Fatalf("disabled terminal was counted: %v %v", terminal, err)
	}
}

func TestValidateOverridesAcceptsOnlyKnownDisabledPlugins(t *testing.T) {
	r := DefaultRegistry()
	for _, raw := range []string{
		`{"rate-limit":null,"circuit-breaker":{}}`,
		`{"static-response":null,"request-id":{}}`,
	} {
		var values map[string]json.RawMessage
		if err := json.Unmarshal([]byte(raw), &values); err != nil {
			t.Fatal(err)
		}
		if err := r.ValidateOverrides(values); err != nil {
			t.Fatalf("valid inheritance override rejected: %s: %v", raw, err)
		}
		if err := r.Validate(values); err == nil {
			t.Fatalf("unresolved null accepted as an effective plugin: %s", raw)
		}
	}
	for _, raw := range []string{
		`{"rate-limt":null}`, `{"unknown":{}}`, `{"rate-limit":{"rate":0,"burst":1}}`, `{"request-id":{"typo":true}}`,
	} {
		var values map[string]json.RawMessage
		_ = json.Unmarshal([]byte(raw), &values)
		if err := r.ValidateOverrides(values); err == nil {
			t.Errorf("invalid override accepted: %s", raw)
		}
	}
	if _, err := r.Terminal(map[string]json.RawMessage{"unknown": []byte(`null`)}); err == nil {
		t.Fatal("unknown disabled plugin accepted by terminal lookup")
	}
}
