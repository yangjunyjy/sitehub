package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"sitehub/common/governance"
)

type Config struct {
	governance.Config
	rest.RestConf
	TrustedProxyCIDRs []string `json:",optional"`
	ProductRPC        zrpc.RpcClientConf
	RPC               zrpc.RpcClientConf
	ServiceKey        string
}
