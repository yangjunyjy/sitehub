package configloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sitehub/common/runtimeconf"

	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type serviceConfig struct {
	rest.RestConf
	RPC        zrpc.RpcClientConf
	Database   runtimeconf.Database
	ServiceKey string
	Password   string        `json:",optional"`
	JobTimeout time.Duration `json:",default=2m"`
}

const validConfig = `Name: test-api
Port: 8802
ServiceKey: ${CONFIG_TEST_SECRET}
Database:
  DSN: ${CONFIG_TEST_DSN}
RPC:
  Endpoints: ["${CONFIG_TEST_RPC}"]
`

func readConfig(t *testing.T, contents, prefix string) (serviceConfig, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	v.SetConfigFile(path)
	var config serviceConfig
	err := Load(v, prefix, &config)
	return config, err
}

func TestLoadPreservesFrameworkDefaultsAndLiteralSecrets(t *testing.T) {
	secret := "quotes\"': # value\nnext-line $CONFIG_TEST_RPC"
	t.Setenv("CONFIG_TEST_SECRET", secret)
	t.Setenv("CONFIG_TEST_DSN", "user:password@tcp(localhost:3306)/blog")
	t.Setenv("CONFIG_TEST_RPC", "127.0.0.1:9802")
	c, err := readConfig(t, validConfig, "LOAD_DEFAULTS")
	if err != nil {
		t.Fatal(err)
	}
	if c.ServiceKey != secret || c.Database.DSN != os.Getenv("CONFIG_TEST_DSN") || c.RPC.Endpoints[0] != "127.0.0.1:9802" {
		t.Fatal("file interpolation must preserve literal environment contents")
	}
	if c.Mode != "pro" || !c.Middlewares.Breaker || !c.Middlewares.Shedding || !c.Middlewares.Timeout || c.CpuThreshold != 900 || c.MaxBytes != 1048576 {
		t.Fatalf("REST resilience defaults were lost: %+v", c.RestConf)
	}
	if !c.RPC.Middlewares.Breaker || !c.RPC.Middlewares.Timeout || !c.RPC.NonBlock || c.RPC.Timeout != 2000 || c.RPC.BalancerName != "p2c_ewma" {
		t.Fatalf("RPC defaults were lost: %+v", c.RPC)
	}
	if c.Database.MaxOpen != 20 || c.Database.Driver != "mysql" || c.JobTimeout != 2*time.Minute {
		t.Fatal("nested and duration defaults were lost")
	}
}

func TestLoadEnvironmentOverridesAbsentFieldsAndPreservesFalse(t *testing.T) {
	t.Setenv("LOAD_ENV_PORT", "8900")
	t.Setenv("LOAD_ENV_MIDDLEWARES_BREAKER", "false")
	t.Setenv("LOAD_ENV_RPC_TIMEOUT", "4567")
	t.Setenv("LOAD_ENV_RPC_NONBLOCK", "false")
	t.Setenv("LOAD_ENV_RPC_ENDPOINTS", `["first:9802","second:9802"]`)
	t.Setenv("LOAD_ENV_DATABASE_MAXOPEN", "7")
	t.Setenv("LOAD_ENV_JOBTIMEOUT", "3m")
	t.Setenv("LOAD_ENV_PASSWORD", "")
	t.Setenv("LOAD_ENV_SERVICEKEY", "literal-$CONFIG_TEST_RPC")
	t.Setenv("CONFIG_TEST_RPC", "must-not-expand")
	t.Setenv("PORT", "9999")
	c, err := readConfig(t, validConfig+"Password: from-file\n", "LOAD_ENV")
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != 8900 || c.Middlewares.Breaker || c.RPC.Timeout != 4567 || c.RPC.NonBlock || c.Database.MaxOpen != 7 || c.JobTimeout != 3*time.Minute {
		t.Fatal("typed environment override was lost")
	}
	if len(c.RPC.Endpoints) != 2 || c.RPC.Endpoints[1] != "second:9802" || c.Password != "" || c.ServiceKey != "literal-$CONFIG_TEST_RPC" {
		t.Fatal("collection, empty or literal environment override was lost")
	}
	other, err := readConfig(t, validConfig, "LOAD_OTHER")
	if err != nil || other.Port != 8802 || !other.Middlewares.Breaker {
		t.Fatal("configuration leaked between service prefixes")
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	for _, test := range []struct {
		name, config string
	}{
		{"missing-required-port", strings.ReplaceAll(validConfig, "Port: 8802\n", "")},
		{"invalid-mode", validConfig + "Mode: invalid\n"},
		{"invalid-range", validConfig + "CpuThreshold: 1001\n"},
		{"invalid-yaml", "Name: [broken\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readConfig(t, test.config, "LOAD_INVALID"); err == nil {
				t.Fatal("invalid configuration was accepted")
			}
		})
	}
	t.Run("invalid-environment-type", func(t *testing.T) {
		t.Setenv("LOAD_INVALID_PORT", "not-a-number")
		if _, err := readConfig(t, validConfig, "LOAD_INVALID"); err == nil || !strings.Contains(err.Error(), "port") {
			t.Fatal("invalid environment port was accepted")
		}
	})
}

func TestLoadRPCServerDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rpc.json")
	if err := os.WriteFile(path, []byte(`{"Name":"test-rpc","ListenOn":"127.0.0.1:9802"}`), 0600); err != nil {
		t.Fatal(err)
	}
	v := viper.New()
	v.SetConfigFile(path)
	var c zrpc.RpcServerConf
	if err := Load(v, "LOAD_RPC", &c); err != nil {
		t.Fatal(err)
	}
	if !c.Middlewares.Breaker || !c.Health || c.Timeout != 2000 || c.CpuThreshold != 900 || c.Mode != "pro" {
		t.Fatalf("RPC server defaults were lost: %+v", c)
	}
}
