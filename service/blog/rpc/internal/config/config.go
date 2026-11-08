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
	RbacRPC    zrpc.RpcClientConf
	FileRPC    zrpc.RpcClientConf
}
