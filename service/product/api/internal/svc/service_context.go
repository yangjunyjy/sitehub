package svc

import (
	"sitehub/service/product/api/internal/config"
	pb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	RPC    pb.ProductClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	c.RPC.Middlewares.Duration = false
	return &ServiceContext{Config: c, RPC: pb.NewProductClient(zrpc.MustNewClient(c.RPC).Conn())}
}
