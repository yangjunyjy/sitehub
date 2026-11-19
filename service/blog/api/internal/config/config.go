package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"sitehub/common/governance"
	"sitehub/common/runtimeconf"
	"sitehub/service/blog/jobs"
)

type Config struct {
	governance.Config
	rest.RestConf
	TrustedProxyCIDRs []string `json:",optional"`
	ProductURL        string
	Database          runtimeconf.Database
	FileDatabase      runtimeconf.Database
	ServiceKey        string
	RbacRPC           zrpc.RpcClientConf
	StorageRoot       string
	StorageSecret     string
	PublicBaseURL     string `json:",optional"`
	Scheduler         jobs.Config
}
