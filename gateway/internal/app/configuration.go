package app

import (
	"crypto/tls"
	"fmt"

	"sitehub/gateway/internal/config"
	"sitehub/gateway/manifest"
	"sitehub/gateway/plugin"
)

// Check also validates startup-only TLS material. Route reloads deliberately do
// not reload listener settings or certificate files.
func Check(c config.Config, plugins *plugin.Registry) (manifest.Document, error) {
	if err := c.Validate(); err != nil {
		return manifest.Document{}, err
	}
	if c.TLS.Certificate != "" {
		if _, err := tls.LoadX509KeyPair(c.TLS.Certificate, c.TLS.Key); err != nil {
			return manifest.Document{}, fmt.Errorf("load gateway TLS certificate: %w", err)
		}
	}
	return ReadRoutes(c, plugins)
}

// ReadRoutes checks file schemas and process capabilities without connecting to
// upstreams, service discovery or the control store. All entry points use it.
func ReadRoutes(c config.Config, plugins *plugin.Registry) (manifest.Document, error) {
	document, err := manifest.ReadWithRegistry(c.Manifest, plugins)
	if err != nil {
		return document, err
	}
	if !c.Discovery.Enabled {
		for name, upstream := range document.Upstreams {
			if upstream.Service != "" {
				return document, fmt.Errorf("upstream %s uses service discovery; enable Discovery or configure static nodes", name)
			}
		}
	}
	return document, nil
}
