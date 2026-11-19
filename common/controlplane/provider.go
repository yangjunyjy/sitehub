package controlplane

import (
	"sitehub/common/configcenter"
	adapter "sitehub/common/controlplane/etcd"
	"sitehub/common/registry"
)

// NewRegistry is the composition point for provider selection. Registering,
// discovery and RPC code consume only the returned proxy's public contract.
func NewRegistry(c Config) (*registry.Proxy, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	backend, err := adapter.NewRegistry(options(c, c.RegistryPrefix()))
	if err != nil {
		return nil, err
	}
	return registry.NewProxy(backend), nil
}

func NewConfigCenter(c Config) (*configcenter.Proxy, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	backend, err := adapter.NewConfigCenter(options(c, c.ConfigPrefix()))
	if err != nil {
		return nil, err
	}
	return configcenter.NewProxy(backend), nil
}

func options(c Config, prefix string) adapter.Options {
	return adapter.Options{Endpoints: c.Endpoints, Username: c.Username, Password: c.Password,
		Prefix: prefix, DialTimeout: c.DialTimeout, RequestTimeout: c.RequestTimeout, LeaseTTL: c.LeaseTTL}
}
