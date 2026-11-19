package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"

	"sitehub/common/governance"
	"sitehub/common/runtimeconf"
)

type Config struct {
	governance.Config
	rest.RestConf
	TrustedProxyCIDRs []string `json:",optional"`
	Database          runtimeconf.Database
	RbacRPC           zrpc.RpcClientConf
	BlogURL           string
	ServiceKey        string
}
