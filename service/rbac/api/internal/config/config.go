package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"sitehub/common/governance"
	"sitehub/service/rbac/api/internal/session"
)

type Config struct {
	BrowserSession session.Config `json:",optional"`
	governance.Config
	rest.RestConf
	TrustedProxyCIDRs []string `json:",optional"`
	RPC               zrpc.RpcClientConf
	ServiceKey        string
}
