package command

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/service"
)

type testConfig struct{ service.ServiceConf }

func TestHelpAndInvalidArgumentsNeverStartService(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		fail bool
	}{
		{"help", []string{"--help"}, false},
		{"unknown-flag", []string{"--unknown"}, true},
		{"positional-argument", []string{"unexpected"}, true},
		{"missing-config", []string{"--config", "missing.yaml"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			started := false
			cmd := New("test-api", "missing.yaml", func(context.Context, testConfig) error { started = true; return nil })
			var output bytes.Buffer
			cmd.SetOut(&output)
			cmd.SetErr(&output)
			cmd.SetArgs(test.args)
			err := cmd.Execute()
			if (err != nil) != test.fail || started {
				t.Fatalf("err=%v, started=%v", err, started)
			}
			if test.name == "help" && !strings.Contains(output.String(), "--config") {
				t.Fatal("Cobra help did not describe configuration flag")
			}
		})
	}
}

func TestConfigAndModePrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	if err := os.WriteFile(path, []byte("Name: test\nMode: test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, env, want string
		args            []string
	}{
		{"file", "", "test", []string{"-f", path}},
		{"environment", "pre", "pre", []string{"--config", path}},
		{"flag", "pre", "dev", []string{"--config", path, "--mode", "dev"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			// Use a distinct service name to avoid global environment interference.
			name := "cmd-" + test.name
			if test.env != "" {
				t.Setenv("CMD_"+strings.ToUpper(test.name)+"_MODE", test.env)
			}
			var got testConfig
			cmd := New(name, "missing.yaml", func(_ context.Context, c testConfig) error { got = c; return nil })
			cmd.SetArgs(test.args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			if got.Mode != test.want {
				t.Fatalf("mode=%s, want %s", got.Mode, test.want)
			}
		})
	}
}

func TestRunReceivesContextAndReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	if err := os.WriteFile(path, []byte("Name: test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	want := errors.New("service failed")
	cmd := New("cmd-context", path, func(got context.Context, c testConfig) error {
		if got.Err() != context.Canceled || c.Mode != "pro" {
			t.Fatal("command context or default mode was lost")
		}
		return want
	})
	cmd.SetArgs([]string{})
	if err := cmd.ExecuteContext(ctx); !errors.Is(err, want) {
		t.Fatalf("service error was lost: %v", err)
	}
}
