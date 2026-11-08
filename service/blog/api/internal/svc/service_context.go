package svc

import (
	"sitehub/service/blog/api/internal/config"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	RPC    pb.BlogClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	if len(c.ServiceKey) < 32 {
		panic("ServiceKey must contain at least 32 characters")
	}
	c.RPC.Middlewares.Duration = false
	return &ServiceContext{Config: c, RPC: pb.NewBlogClient(zrpc.MustNewClient(c.RPC).Conn())}
}
