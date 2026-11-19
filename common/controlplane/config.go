// Package controlplane assembles provider adapters behind the registry and
// configuration-center proxies. Consumers never import a provider SDK.
package controlplane

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Provider       string        `json:",default=etcd"`
	Endpoints      []string      `json:",optional"`
	Prefix         string        `json:",default=/sitehub"`
	Namespace      string        `json:",default=sitehub-dev"`
	Group          string        `json:",default=SITEHUB"`
	Username       string        `json:",optional"`
	Password       string        `json:",optional"`
	DialTimeout    time.Duration `json:",default=5s"`
	RequestTimeout time.Duration `json:",default=5s"`
	LeaseTTL       int64         `json:",default=15"`
}

func (c Config) Validate() error {
	if c.Provider != "etcd" {
		return fmt.Errorf("unsupported control-plane provider")
	}
	if len(c.Endpoints) == 0 {
		return fmt.Errorf("ControlPlane.Endpoints are required")
	}
	for _, value := range c.Endpoints {
		address := value
		if strings.Contains(value, "://") {
			u, err := url.Parse(value)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
				return fmt.Errorf("invalid control-plane endpoint")
			}
			address = u.Host
		}
		host, port, err := net.SplitHostPort(address)
		n, portErr := strconv.ParseUint(port, 10, 16)
		if err != nil || host == "" || portErr != nil || n == 0 {
			return fmt.Errorf("control-plane endpoints must contain host:port")
		}
	}
	if !strings.HasPrefix(c.Prefix, "/") || path.Clean(c.Prefix) != c.Prefix || c.Prefix == "/" {
		return fmt.Errorf("ControlPlane.Prefix must be a non-root absolute key prefix")
	}
	for _, value := range []string{c.Namespace, c.Group} {
		if !validSegment(value) {
			return fmt.Errorf("control-plane namespace and group must be non-empty path segments")
		}
	}
	if c.DialTimeout <= 0 || c.RequestTimeout <= 0 || c.RequestTimeout > time.Minute || c.LeaseTTL < 5 {
		return fmt.Errorf("control-plane timeouts must be positive, RequestTimeout <= 1m and LeaseTTL >= 5s")
	}
	return nil
}

func validSegment(value string) bool {
	return value != "" && value != "." && value != ".." && !strings.ContainsAny(value, "/\\\x00")
}

func (c Config) RegistryPrefix() string {
	return path.Join(c.Prefix, "registry", c.Namespace, c.Group) + "/"
}
func (c Config) ConfigPrefix() string {
	return path.Join(c.Prefix, "config", c.Namespace, c.Group) + "/"
}
