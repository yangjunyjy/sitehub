package svc

import (
	"sitehub/service/rbac/api/internal/config"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	RPC    pb.RbacClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	c.RPC.Middlewares.Duration = false
	return &ServiceContext{Config: c, RPC: pb.NewRbacClient(zrpc.MustNewClient(c.RPC).Conn())}
}
