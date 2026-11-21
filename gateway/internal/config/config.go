package config

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/common/controlplane"
	"sitehub/gateway/console"
)

type Config struct {
	Name   string `json:",default=gateway"`
	Listen string `json:",default=0.0.0.0:9080"`
	// Manifest points to declarative proxy configuration, never business services.
	Manifest string `json:",default=routes.yaml"`
	// Source selects file reloads or shared etcd publication. An omitted value
	// preserves existing deployments with Etcd.Endpoints; otherwise it uses files.
	Source string `json:",optional"`
	Log    logx.LogConf
	Admin  struct {
		Listen       string `json:",default=127.0.0.1:9180"`
		Key          string
		AllowedCIDRs []string
	}
	Etcd struct {
		Endpoints []string
		Key       string `json:",default=/sitehub/gateway/v1/config"`
		Username  string `json:",optional"`
		Password  string `json:",optional"`
	} `json:",optional"`
	Console struct {
		console.Config
		Enabled  bool   `json:",default=false"`
		StateDir string `json:",default=.tmp/gateway-releases"`
	} `json:",optional"`
	ControlPlane controlplane.Config `json:",optional"`
	Discovery    struct {
		Enabled  bool          `json:",default=false"`
		MaxStale time.Duration `json:",default=30s"`
	}
	HTTP struct {
		MaxBodyBytes      int64         `json:",default=2149580800"`
		ReadHeaderTimeout time.Duration `json:",default=10s"`
		IdleTimeout       time.Duration `json:",default=90s"`
		ShutdownTimeout   time.Duration `json:",default=60s"`
	}
	TLS struct {
		Certificate string `json:",optional"`
		Key         string `json:",optional"`
	}
}

func (c Config) SourceType() string {
	if c.Source != "" {
		return c.Source
	}
	if len(c.Etcd.Endpoints) > 0 {
		return "etcd"
	}
	return "file"
}

func (c Config) Validate() error {
	addresses := []string{c.Listen, c.Admin.Listen}
	if c.Console.Enabled {
		if err := c.Console.Validate(); err != nil {
			return err
		}
		addresses = append(addresses, c.Console.Listen)
	}
	for _, address := range addresses {
		_, port, err := net.SplitHostPort(address)
		number, parseErr := strconv.Atoi(port)
		if err != nil || parseErr != nil || number < 1 || number > 65535 {
			return fmt.Errorf("Listen and Admin.Listen must be host:port with a port between 1 and 65535")
		}
	}
	if c.Listen == c.Admin.Listen || len(c.Admin.Key) < 32 || c.Manifest == "" {
		return fmt.Errorf("distinct listeners, Admin.Key >=32 characters and Manifest are required")
	}
	if c.SourceType() != "file" && c.SourceType() != "etcd" {
		return fmt.Errorf("Source must be file or etcd")
	}
	if c.SourceType() == "etcd" && (len(c.Etcd.Endpoints) == 0 || c.Etcd.Key == "") {
		return fmt.Errorf("etcd source requires Etcd.Endpoints and Etcd.Key")
	}
	if len(c.Admin.AllowedCIDRs) == 0 {
		return fmt.Errorf("Admin.AllowedCIDRs cannot be empty")
	}
	for _, value := range c.Admin.AllowedCIDRs {
		if _, _, err := net.ParseCIDR(value); err != nil {
			return fmt.Errorf("invalid Admin.AllowedCIDRs")
		}
	}
	if c.Discovery.Enabled {
		if err := c.ControlPlane.Validate(); err != nil {
			return err
		}
	}
	if c.HTTP.MaxBodyBytes <= 0 || c.HTTP.ReadHeaderTimeout <= 0 || c.HTTP.IdleTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 || c.Discovery.MaxStale <= 0 {
		return fmt.Errorf("HTTP limits and Discovery.MaxStale must be positive")
	}
	if (c.TLS.Certificate == "") != (c.TLS.Key == "") {
		return fmt.Errorf("TLS certificate and key must be configured together")
	}
	return nil
}
