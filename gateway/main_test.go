package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckResolvesFilesRelativeToProcessConfigWithoutInfrastructure(t *testing.T) {
	root := t.TempDir()
	configuration := "Source: file\nManifest: routes.yaml\nAdmin:\n  Key: test-key-with-at-least-32-characters\n  AllowedCIDRs: [127.0.0.1/32]\nDiscovery:\n  Enabled: false\n"
	routes := "version: 1\nroutes:\n  - id: ping\n    host: example.test\n    paths: [/ping]\n    plugins:\n      static-response: {status: 200, body: pong}\n"
	for name, content := range map[string]string{"gateway.yaml": configuration, "routes.yaml": routes} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	command := newCommand()
	var output bytes.Buffer
	command.SetOut(&output)
	command.SetArgs([]string{"check", "-f", filepath.Join(root, "gateway.yaml")})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "are valid") {
		t.Fatal("configuration was not checked")
	}
	for _, invalid := range []string{
		configuration + "Listen: 127.0.0.1:notaport\n",
		configuration + "TLS:\n  Certificate: missing-cert.pem\n  Key: missing-key.pem\n",
	} {
		if err := os.WriteFile(filepath.Join(root, "gateway.yaml"), []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		command = newCommand()
		command.SetArgs([]string{"check", "-f", filepath.Join(root, "gateway.yaml")})
		if err := command.Execute(); err == nil {
			t.Fatal("check accepted invalid listener or missing TLS material")
		}
	}
	if err := os.WriteFile(filepath.Join(root, "gateway.yaml"), []byte(configuration), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "routes.yaml"), []byte(routes+"unknown: true\n"), 0600); err != nil {
		t.Fatal(err)
	}
	command = newCommand()
	command.SetArgs([]string{"check", "-f", filepath.Join(root, "gateway.yaml")})
	if err := command.Execute(); err == nil {
		t.Fatal("check accepted an unknown proxy setting")
	}
}
