package governance

import (
	center "sitehub/common/configcenter/nacos"
	"sitehub/common/nacosconn"
	registry "sitehub/common/registry/nacos"
	"sitehub/common/runtimeconfig"
)

// Config is embedded in each service config. Nacos and ConfigCenter remain
// local bootstrap settings; only Runtime fields are live-reloadable.
type Config struct {
	Nacos        nacosconn.Config     `json:",optional"`
	Registry     registry.Config      `json:",optional"`
	ConfigCenter center.Config        `json:",optional"`
	Runtime      runtimeconfig.Config `json:",optional"`
	Management   struct {
		ListenOn string `json:",optional"`
	} `json:",optional"`
}

func (c Config) GovernanceConfig() Config { return c }
