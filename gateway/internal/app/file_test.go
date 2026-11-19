package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sitehub/common/configloader"
	"sitehub/gateway/internal/config"
	"sitehub/gateway/plugin"

	"github.com/spf13/viper"
)

func TestFileGatewayUsesCustomRegistryAndReloadsWithoutExternalServices(t *testing.T) {
	port := func() string {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()
		return listener.Addr().String()
	}
	public, admin := port(), port()
	root := t.TempDir()
	routes := filepath.Join(root, "routes.yaml")
	writeRoutes := func(body string) {
		t.Helper()
		content := "version: 1\nroutes:\n  - id: custom\n    host: example.test\n    paths: [/readyz]\n    plugins:\n      custom-response: {body: " + body + "}\n"
		if err := os.WriteFile(routes, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeRoutes("old")
	bootstrap := filepath.Join(root, "gateway.yaml")
	key := "file-test-control-key-at-least-32-characters"
	content := fmt.Sprintf("Source: file\nListen: %s\nManifest: %s\nAdmin:\n  Listen: %s\n  Key: %s\n  AllowedCIDRs: [127.0.0.1/32]\nDiscovery:\n  Enabled: false\nLog:\n  Mode: console\n  Level: error\nHTTP:\n  ShutdownTimeout: 1s\n", public, filepath.ToSlash(routes), admin, key)
	if err := os.WriteFile(bootstrap, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	settings := viper.New()
	settings.SetConfigFile(bootstrap)
	var c config.Config
	if err := configloader.Load(settings, "FILE_GATEWAY_TEST", &c); err != nil {
		t.Fatal(err)
	}
	plugins := plugin.DefaultRegistry()
	if err := plugins.RegisterTerminal("custom-response", 100, func(raw json.RawMessage) (plugin.Middleware, error) {
		var value struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return func(http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, value.Body) })
		}, nil
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan error, 1)
	go func() { finished <- RunWithRegistry(ctx, c, plugins) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-finished:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(5 * time.Second):
			t.Error("file gateway did not stop")
		}
	})
	client := &http.Client{Timeout: time.Second}
	ready := false
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get("http://" + admin + "/readyz")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !ready {
		t.Fatal("file gateway failed to become ready")
	}
	request := func(method, address string) (int, string) {
		t.Helper()
		req, err := http.NewRequest(method, address, nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Host = "example.test"
		req.Header.Set("Authorization", "Bearer "+key)
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		return response.StatusCode, string(body)
	}
	if status, body := request("GET", "http://"+public+"/readyz"); status != 200 || body != "old" {
		t.Fatalf("custom route blocked: %d %s", status, body)
	}
	writeRoutes("new")
	if status, body := request("POST", "http://"+admin+"/v1/reload"); status != 200 {
		t.Fatalf("reload failed: %d %s", status, body)
	}
	if _, body := request("GET", "http://"+public+"/readyz"); body != "new" {
		t.Fatal("file reload did not switch route")
	}
	if err := os.WriteFile(routes, []byte("version: 1\nunknown: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if status, _ := request("POST", "http://"+admin+"/v1/reload"); status != 400 {
		t.Fatal("invalid reload accepted")
	}
	if _, body := request("GET", "http://"+public+"/readyz"); body != "new" {
		t.Fatal("invalid reload changed active route")
	}
}

func TestCheckRejectsDiscoveryUpstreamWhenDiscoveryDisabled(t *testing.T) {
	file := filepath.Join(t.TempDir(), "routes.yaml")
	content := "version: 1\nupstreams:\n  orders:\n    service: Orders.Service\nroutes:\n  - id: orders\n    host: orders.test\n    paths: ['/*']\n    upstream: orders\n"
	if err := os.WriteFile(file, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	c := config.Config{Manifest: file}
	if _, err := ReadRoutes(c, plugin.DefaultRegistry()); err == nil || !strings.Contains(err.Error(), "enable Discovery") {
		t.Fatalf("discovery prerequisite missed: %v", err)
	}
}
