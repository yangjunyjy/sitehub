package config

import (
	"fmt"
	"net"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/common/nacosconn"
)

type Config struct {
	Name     string `json:",default=gateway"`
	Listen   string `json:",default=0.0.0.0:9080"`
	Manifest string `json:",default=gateway/etc/routes.json"`
	Log      logx.LogConf
	Admin    struct {
		Listen       string `json:",default=127.0.0.1:9180"`
		Key          string
		AllowedCIDRs []string
	}
	Etcd struct {
		Endpoints []string
		Key       string `json:",default=/sitehub/gateway/v1/config"`
		Username  string `json:",optional"`
		Password  string `json:",optional"`
	}
	Nacos     nacosconn.Config
	Discovery struct {
		Enabled  bool          `json:",default=true"`
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

func (c Config) Validate() error {
	for _, address := range []string{c.Listen, c.Admin.Listen} {
		if _, _, err := net.SplitHostPort(address); err != nil {
			return fmt.Errorf("Listen and Admin.Listen must be host:port")
		}
	}
	if c.Listen == c.Admin.Listen || len(c.Admin.Key) < 32 || len(c.Etcd.Endpoints) == 0 || c.Etcd.Key == "" {
		return fmt.Errorf("distinct listeners, Admin.Key >=32 characters and Etcd configuration are required")
	}
	if c.HTTP.MaxBodyBytes <= 0 || c.HTTP.ReadHeaderTimeout <= 0 || c.HTTP.IdleTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 || c.Discovery.MaxStale <= 0 {
		return fmt.Errorf("HTTP limits and Discovery.MaxStale must be positive")
	}
	if (c.TLS.Certificate == "") != (c.TLS.Key == "") {
		return fmt.Errorf("TLS certificate and key must be configured together")
	}
	return nil
}
