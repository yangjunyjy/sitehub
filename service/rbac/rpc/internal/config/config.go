package config

import (
	"sitehub/common/governance"
	"sitehub/common/runtimeconf"

	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	governance.Config
	zrpc.RpcServerConf
	Database   runtimeconf.Database
	ServiceKey string
	Cache      struct {
		Addr     string
		Password string `json:",optional"`
		DB       int    `json:",default=0"`
	}
	SessionTTL   int `json:",default=86400"`
	OnceTokenTTL int `json:",default=300"`
	AuthSecret   string
	TOTP         struct {
		Enabled    bool   `json:",default=false"`
		Secret     string `json:",optional"`
		TimeStep   int    `json:",default=30"`
		WindowSize int    `json:",default=1"`
	}
	Mail struct {
		Host     string `json:",optional"`
		Port     int    `json:",default=465"`
		Username string `json:",optional"`
		Password string `json:",optional"`
		FromName string `json:",default=SiteHub"`
	}
	BlacklistEmail string `json:",optional"`
}
