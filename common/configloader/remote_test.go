package configloader

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemotePrecedenceAndBootstrapProtection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	_ = os.WriteFile(path, []byte("Name: local\nPort: 80\n"), 0600)
	type config struct {
		Name string
		Port int
	}
	fresh := func() *viper.Viper { v := viper.New(); v.SetConfigFile(path); return v }
	var c config
	if err := LoadRemote(fresh(), "REMOTE_TEST", &c, []byte("Name: remote\nPort: 81\n")); err != nil || c.Port != 81 {
		t.Fatalf("remote: %+v %v", c, err)
	}
	t.Setenv("REMOTE_TEST_PORT", "82")
	if err := LoadRemote(fresh(), "REMOTE_TEST", &c, []byte("Name: remote\nPort: 81\n")); err != nil || c.Port != 82 {
		t.Fatalf("environment: %+v %v", c, err)
	}
	v := fresh()
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("port", 0, "")
	_ = flags.Set("port", "83")
	_ = v.BindPFlag("Port", flags.Lookup("port"))
	if err := LoadRemote(v, "REMOTE_TEST", &c, []byte("Name: remote\nPort: 81\n")); err != nil || c.Port != 83 {
		t.Fatalf("flag: %+v %v", c, err)
	}
	if err := LoadRemote(fresh(), "REMOTE_TEST", &c, []byte("ControlPlane:\n  Password: forbidden\n")); err == nil {
		t.Fatal("remote bootstrap override accepted")
	}
}

func TestRemoteCannotRepointBootstrap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.yaml")
	if err := os.WriteFile(path, []byte("Name: local\nControlPlane:\n  Password: local\nConfigCenter:\n  Key: local.yaml\n"), 0600); err != nil {
		t.Fatal(err)
	}
	type config struct {
		Name         string
		ControlPlane struct{ Password string }
		ConfigCenter struct{ Key string }
	}
	for _, remote := range []string{"ControlPlane:\n  Password: remote\n", "ConfigCenter:\n  Key: other.yaml\n"} {
		settings := viper.New()
		settings.SetConfigFile(path)
		var c config
		if err := LoadRemote(settings, "BOOTSTRAP_TEST", &c, []byte(remote)); err == nil || !strings.Contains(err.Error(), "remote configuration cannot set") {
			t.Fatalf("known bootstrap field must be rejected by bootstrap protection, got %v", err)
		}
	}
}
