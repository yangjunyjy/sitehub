package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"sitehub/gateway/plugin"
)

func TestEffectiveProjectConfiguration(t *testing.T) {
	jsonDocument, err := Read(filepath.Join("..", "etc", "routes.json"))
	if err != nil {
		t.Fatal(err)
	}
	yamlDocument, err := Read(filepath.Join("..", "etc", "routes.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	// Route array order does not affect matching: the engine orders routes by
	// priority and ID. Plugin object key order is likewise not significant.
	// Normalize only in this comparison; persisted revision hashes stay intact.
	canonical := func(document Document) []byte {
		document.Routes = append([]Route(nil), document.Routes...)
		sort.Slice(document.Routes, func(i, j int) bool { return document.Routes[i].ID < document.Routes[j].ID })
		var value any
		decoder := json.NewDecoder(bytes.NewReader(Bytes(document)))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	if !bytes.Equal(canonical(jsonDocument), canonical(yamlDocument)) {
		t.Fatal("project YAML includes changed the effective JSON gateway configuration")
	}
}

func writeConfig(t *testing.T, root, name, content string) string {
	t.Helper()
	file := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

const staticYAML = `version: 1
routes:
  - id: root
    host: example.test
    paths: ["/*"]
    plugins:
      static-response: {status: 200, body: ok}
`

func TestIncludesResolveRelativePathsAndApplyPluginPrecedence(t *testing.T) {
	root := t.TempDir()
	file := writeConfig(t, root, "routes.yaml", `version: 1
plugins:
  request-id: {}
  rate-limit: {rate: 20, burst: 20}
  circuit-breaker: {}
include: [conf.d/*.yaml]
`)
	writeConfig(t, root, "conf.d/20-routes.yaml", `include: [nested/more.yml]
routes:
  - id: shop
    host: shop.example.test
    paths: ["/*"]
    upstream: inventory
    profile: browser
    plugins:
      request-id: {}
      rate-limit: null
`)
	writeConfig(t, root, "conf.d/nested/more.yml", `routes:
  - id: second
    host: status.example.test
    paths: ["/*"]
    plugins:
      static-response: {status: 200, body: healthy}
`)
	writeConfig(t, root, "conf.d/10-common.yaml", `upstreams:
  inventory:
    service: inventory.v2
profiles:
  browser:
    request-id: null
    rate-limit: {rate: 40, burst: 40}
`)
	document, err := Read(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Routes) != 2 || document.Routes[0].ID != "shop" || document.Routes[1].ID != "second" {
		t.Fatalf("unexpected include order: %+v", document.Routes)
	}
	upstream := document.Upstreams["inventory"]
	if upstream.Service != "inventory.v2" || upstream.ConnectTimeoutMS != 3000 || upstream.ResponseHeaderTimeoutMS != 10000 || upstream.IOTimeoutMS != 30000 || upstream.HealthPath != "/healthz" || upstream.HealthIntervalMS != 2000 || upstream.HealthTimeoutMS != 1000 || upstream.HealthFailures != 2 {
		t.Fatalf("file defaults were not resolved: %+v", upstream)
	}
	effective := EffectivePlugins(document, document.Routes[0])
	if _, exists := effective["rate-limit"]; exists {
		t.Fatal("route could not disable inherited rate-limit")
	}
	if _, exists := effective["request-id"]; !exists {
		t.Fatal("route could not re-enable request-id disabled by its profile")
	}
	if _, exists := effective["circuit-breaker"]; !exists {
		t.Fatal("global plugin was not inherited")
	}
	if strings.Contains(string(Bytes(document)), `"include"`) {
		t.Fatal("file includes escaped into the published document")
	}
	decoded, err := Decode(Bytes(document))
	if err != nil || Validate(decoded, plugin.DefaultRegistry()) != nil || Revision(decoded) != Revision(document) {
		t.Fatalf("resolved document did not round trip: %v", err)
	}
}

func TestFileDefaultsDoNotChangeWireValidation(t *testing.T) {
	raw := `{"version":1,"upstreams":{"app":{"nodes":[{"url":"http://127.0.0.1:9001"}]}},"routes":[{"id":"app","host":"example.test","paths":["/*"],"upstream":"app"}]}`
	document, err := Decode([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(document, plugin.DefaultRegistry()); err == nil {
		t.Fatal("wire input was allowed to omit explicit runtime defaults")
	}
	file := writeConfig(t, t.TempDir(), "routes.json", raw)
	document, err = Read(file)
	if err != nil {
		t.Fatal(err)
	}
	if document.Upstreams["app"].Nodes[0].Weight != 1 {
		t.Fatal("missing node weight did not default to one")
	}
	for _, field := range []string{`"weight":0,`, `"weight":null,`} {
		file := writeConfig(t, t.TempDir(), "routes.json", strings.Replace(raw, `"url":`, field+`"url":`, 1))
		if _, err := Read(file); err == nil {
			t.Fatalf("explicit invalid node weight accepted: %s", field)
		}
	}
}

func TestDecodeRejectsDuplicateWireFields(t *testing.T) {
	for _, raw := range []string{
		`{"version":1,"version":2}`,
		`{"version":1,"plugins":{"request-id":{},"request-id":null}}`,
	} {
		if _, err := Decode([]byte(raw)); err == nil || !strings.Contains(err.Error(), "duplicate JSON key") {
			t.Fatalf("ambiguous published JSON accepted: %s, %v", raw, err)
		}
	}
}

func TestReadRejectsAmbiguousAndInvalidFiles(t *testing.T) {
	for _, test := range []struct {
		name, extension, content, want string
	}{
		{"unknown-root", ".yaml", "version: 1\nroutez: []\n", "unknown field"},
		{"unknown-route", ".yaml", strings.Replace(staticYAML, "id: root", "id: root\n    typo: true", 1), "unknown field"},
		{"duplicate-yaml", ".yaml", "version: 1\nversion: 1\n", "duplicate key"},
		{"duplicate-json", ".json", `{"version":1,"version":1}`, "duplicate JSON key"},
		{"nested-json-duplicate", ".json", `{"version":1,"plugins":{"request-id":{},"request-id":{}}}`, "duplicate JSON key"},
		{"multiple-yaml", ".yml", staticYAML + "---\nversion: 1\n", "one YAML document"},
		{"multiple-json", ".json", `{"version":1} {"version":1}`, "one JSON document"},
		{"missing-version", ".yaml", strings.TrimPrefix(staticYAML, "version: 1\n"), "root version must be 1"},
		{"unsupported-version", ".yaml", strings.Replace(staticYAML, "version: 1", "version: 2", 1), "root version must be 1"},
		{"unknown-disabled-global", ".yaml", "plugins: {unknown-plugin: null}\n" + staticYAML, "unknown plugin"},
		{"unknown-disabled-route", ".yaml", strings.Replace(staticYAML, "static-response:", "unknown-plugin: null\n      static-response:", 1), "unknown plugin"},
		{"non-string-key", ".yaml", "version: 1\nplugins: {1: {}}\n", "keys must be strings"},
		{"alias", ".yaml", "version: 1\nplugins: &p {}\nprofiles: {one: *p}\n", "aliases are not supported"},
		{"merge", ".yaml", "version: 1\nplugins: {<<: {request-id: {}}}\n", "merge keys are not supported"},
		{"unknown-upstream", ".yaml", "version: 1\nupstreams: {app: {service: inventory, typo: true}}\n", "unknown field"},
		{"unknown-plugin-field", ".yaml", strings.Replace(staticYAML, "{status: 200, body: ok}", "{status: 200, typo: true}", 1), "unknown field"},
		{"empty-timeout", ".yaml", "version: 1\nupstreams: {app: {service: inventory, connect_timeout_ms: 0}}\nroutes: [{id: app, host: example.test, paths: ['/*'], upstream: app}]\n", "invalid timeout"},
	} {
		t.Run(test.name, func(t *testing.T) {
			file := writeConfig(t, t.TempDir(), "routes"+test.extension, test.content)
			_, err := Read(file)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
		})
	}
}

func TestReadRejectsBrokenIncludesAndDuplicateDefinitions(t *testing.T) {
	for _, test := range []struct {
		name, root, included, want string
	}{
		{"missing", "version: 1\ninclude: [missing/*.yaml]\n", "", "matched no files"},
		{"cycle", "version: 1\ninclude: [child.yaml]\n", "include: [routes.yaml]\n", "include cycle"},
		{"repeated-file", "version: 1\ninclude: [child.yaml, child.yaml]\n", "{}\n", "included more than once"},
		{"included-version", "version: 1\ninclude: [child.yaml]\n", "version: 2\n", "included versions"},
		{"duplicate-upstream", "version: 1\ninclude: [child.yaml]\nupstreams: {app: {service: app}}\n", "upstreams: {app: {service: app}}\n", "duplicate upstream"},
		{"duplicate-profile", "version: 1\ninclude: [child.yaml]\nprofiles: {one: {}}\n", "profiles: {one: {}}\n", "duplicate profile"},
		{"duplicate-route", "include: [child.yaml]\n" + staticYAML, strings.TrimPrefix(staticYAML, "version: 1\n"), "duplicate route"},
		{"duplicate-global", "version: 1\ninclude: [child.yaml]\nplugins: {request-id: {}}\n", "plugins: {request-id: {}}\n", "duplicate global plugin"},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			file := writeConfig(t, root, "routes.yaml", test.root)
			if test.included != "" {
				writeConfig(t, root, "child.yaml", test.included)
			}
			_, err := Read(file)
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), ".yaml") {
				t.Fatalf("expected contextual %q, got %v", test.want, err)
			}
			if strings.HasPrefix(test.want, "duplicate") && !strings.Contains(err.Error(), "child.yaml") {
				t.Fatalf("duplicate error omitted second source: %v", err)
			}
		})
	}
}

func TestReadLimitsCombinedFilesAndByteSize(t *testing.T) {
	t.Run("files", func(t *testing.T) {
		root := t.TempDir()
		file := writeConfig(t, root, "routes.yaml", "version: 1\ninclude: [parts/*.yaml]\n")
		for index := 0; index < maxFiles; index++ {
			writeConfig(t, root, fmt.Sprintf("parts/%03d.yaml", index), "{}\n")
		}
		if _, err := Read(file); err == nil || !strings.Contains(err.Error(), "exceed 128 files") {
			t.Fatalf("unbounded file includes: %v", err)
		}
	})
	t.Run("combined-bytes", func(t *testing.T) {
		root := t.TempDir()
		file := writeConfig(t, root, "routes.yaml", "version: 1\ninclude: [child.yaml]\n#"+strings.Repeat("x", maxFileBytes/2))
		writeConfig(t, root, "child.yaml", "{}\n#"+strings.Repeat("x", maxFileBytes/2))
		if _, err := Read(file); err == nil || !strings.Contains(err.Error(), "exceed 4 MiB") {
			t.Fatalf("unbounded file input: %v", err)
		}
	})
}

func TestGlobalPluginsAffectRevisionAndCanBeDisabledByProfile(t *testing.T) {
	file := writeConfig(t, t.TempDir(), "routes.yaml", "plugins: {request-id: {}}\nprofiles: {public: {request-id: null}}\n"+strings.Replace(staticYAML, "id: root", "id: root\n    profile: public", 1))
	document, err := Read(file)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := EffectivePlugins(document, document.Routes[0])["request-id"]; exists {
		t.Fatal("profile could not disable global plugin")
	}
	first := Revision(document)
	document.Plugins["circuit-breaker"] = json.RawMessage(`{}`)
	if first == Revision(document) {
		t.Fatal("revision omitted global plugins")
	}
}

func TestCustomTerminalPluginDoesNotNeedManifestChanges(t *testing.T) {
	registry := plugin.NewRegistry()
	err := registry.RegisterTerminal("custom-reply", 100, func(raw json.RawMessage) (plugin.Middleware, error) {
		var config struct {
			Body string `json:"body"`
		}
		if err := json.Unmarshal(raw, &config); err != nil {
			return nil, err
		}
		return func(http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(config.Body)) })
		}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	file := writeConfig(t, t.TempDir(), "routes.yaml", strings.Replace(staticYAML, "static-response: {status: 200, body: ok}", "custom-reply: {body: custom}", 1))
	if _, err := Read(file); err == nil {
		t.Fatal("built-in registry accepted an unregistered plugin")
	}
	document, err := ReadWithRegistry(file, registry)
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Routes) != 1 || document.Routes[0].Upstream != "" {
		t.Fatal("custom terminal did not load independently of upstreams")
	}
	if _, err := ReadWithRegistry(file, nil); err == nil {
		t.Fatal("nil registry accepted")
	}
}
