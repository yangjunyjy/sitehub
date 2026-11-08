package svc

import (
	"sitehub/service/file/api/internal/config"
	pb "sitehub/service/file/rpc/pb/file"
	productpb "sitehub/service/product/rpc/pb/product"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Product productpb.ProductClient
	Config  config.Config
	RPC     pb.FileClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	c.RPC.Middlewares.Duration = false
	return &ServiceContext{Config: c, Product: productpb.NewProductClient(zrpc.MustNewClient(c.ProductRPC).Conn()), RPC: pb.NewFileClient(zrpc.MustNewClient(c.RPC).Conn())}
}
