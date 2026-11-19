package controlplane

import (
	"strings"
	"testing"
	"time"
)

func configured() Config {
	return Config{Provider: "etcd", Endpoints: []string{"127.0.0.1:2379"}, Prefix: "/sitehub", Namespace: "dev", Group: "SITEHUB", DialTimeout: time.Second, RequestTimeout: time.Second, LeaseTTL: 15}
}

func TestProviderIsolationAndInvalidBootstrap(t *testing.T) {
	c := configured()
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.RegistryPrefix() != "/sitehub/registry/dev/SITEHUB/" || c.ConfigPrefix() != "/sitehub/config/dev/SITEHUB/" {
		t.Fatal("registry and config storage overlap")
	}
	for _, mutate := range []func(*Config){
		func(c *Config) { c.Provider = "missing" },
		func(c *Config) { c.Prefix = "/sitehub/../other" },
		func(c *Config) { c.Namespace = "../other" },
		func(c *Config) { c.Group = "/other" },
		func(c *Config) { c.Endpoints = []string{"http://user:password@localhost:2379"} },
		func(c *Config) { c.Endpoints = []string{"localhost:70000"} },
		func(c *Config) { c.LeaseTTL = 0 },
	} {
		c := configured()
		mutate(&c)
		if err := c.Validate(); err == nil {
			t.Fatalf("invalid bootstrap accepted: provider=%s", c.Provider)
		} else if strings.Contains(err.Error(), "password") {
			t.Fatal("credential leaked")
		}
	}
}
