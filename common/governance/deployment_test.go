package governance

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"sitehub/common/configloader"

	"github.com/spf13/viper"
)

// go-zero omits these fields from ShutdownConf on Windows. Use Linux's
// duration layout so local template checks still enforce the production
// withdrawal window on every development platform.
type cloudShutdownConfig struct {
	Config
	Shutdown struct{ WrapUpTime, WaitTime time.Duration }
}

// These checks deliberately load only the governance part of each remote seed;
// cloudShutdownConfig does not describe service-specific configuration fields.
// The instance shutdown budget remains in the local deployment configuration.
func cloudDeploymentGovernance(t *testing.T, name string) (*viper.Viper, []byte) {
	t.Helper()
	_, source, _, _ := runtime.Caller(0)
	deployDir := filepath.Join(filepath.Dir(source), "..", "..", "deploy")
	content, err := os.ReadFile(filepath.Join(deployDir, "etcd", "config", name+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	seed := viper.New()
	seed.SetConfigType("yaml")
	if err := seed.ReadConfig(bytes.NewReader(content)); err != nil {
		t.Fatal(err)
	}
	governance := make(map[string]any)
	for _, key := range []string{"Registry", "Runtime"} {
		if !seed.IsSet(key) {
			t.Fatalf("remote seed %s must provide %s", name, key)
		}
		governance[key] = seed.Get(key)
	}
	remote, err := json.Marshal(governance)
	if err != nil {
		t.Fatal(err)
	}
	settings := viper.New()
	settings.SetConfigFile(filepath.Join(deployDir, "config", name+".yaml"))
	return settings, remote
}

func TestCloudDeploymentShutdownWindows(t *testing.T) {
	t.Setenv("ETCD_ENDPOINT", "127.0.0.1:2379")
	t.Setenv("CONTROL_NAMESPACE", "sitehub-dev")
	for _, name := range []string{"rbac-api", "rbac-rpc", "blog-api", "product-api"} {
		t.Run(name, func(t *testing.T) {
			var c cloudShutdownConfig
			settings, remote := cloudDeploymentGovernance(t, name)
			if err := configloader.LoadRemote(settings, "CLOUD_SHUTDOWN_TEST", &c, remote); err != nil {
				t.Fatal(err)
			}
			if !c.Registry.Enabled {
				t.Fatal("cloud service must enable registration before testing its shutdown window")
			}
			if err := Validate(c, c.Config); err != nil {
				t.Fatalf("deployment configuration cannot pass Linux startup validation: %v", err)
			}
		})
	}
}

func TestCloudDeploymentRejectsShortShutdownWindow(t *testing.T) {
	t.Setenv("ETCD_ENDPOINT", "127.0.0.1:2379")
	t.Setenv("CONTROL_NAMESPACE", "sitehub-dev")
	settings, remote := cloudDeploymentGovernance(t, "rbac-api")
	// Exercise the same loader with the original invalid deployment values;
	// changing shared template files during a test would race with packaging.
	settings.Set("ControlPlane.RequestTimeout", "5s")
	settings.Set("ControlPlane.LeaseTTL", 15)
	settings.Set("Shutdown.WrapUpTime", "10s")
	var c cloudShutdownConfig
	if err := configloader.LoadRemote(settings, "CLOUD_SHUTDOWN_TEST", &c, remote); err != nil {
		t.Fatal(err)
	}
	if err := Validate(c, c.Config); err == nil || !strings.Contains(err.Error(), "at least 20s") {
		t.Fatalf("the original deployment values must fail Linux shutdown validation, got %v", err)
	}
}
