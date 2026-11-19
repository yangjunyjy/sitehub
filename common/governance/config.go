package governance

import (
	"sitehub/common/configcenter"
	"sitehub/common/controlplane"
	"sitehub/common/registry"
	"sitehub/common/runtimeconfig"
)

// Config is embedded in each service config. ControlPlane and ConfigCenter remain
// local bootstrap settings; only Runtime fields are live-reloadable.
type Config struct {
	ControlPlane controlplane.Config  `json:",optional"`
	Registry     registry.Config      `json:",optional"`
	ConfigCenter configcenter.Config  `json:",optional"`
	Runtime      runtimeconfig.Config `json:",optional"`
	Management   struct {
		ListenOn string `json:",optional"`
	} `json:",optional"`
}

func (c Config) GovernanceConfig() Config { return c }
