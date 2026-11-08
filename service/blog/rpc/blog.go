package main

import (
	"context"
	"sitehub/common/command"
	"sitehub/common/governance"
	"sitehub/common/rpcx"
	"sitehub/service/blog/rpc/internal/config"
	healthlogic "sitehub/service/blog/rpc/internal/logic"
	"sitehub/service/blog/rpc/internal/server"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func main() {
	command.Execute(command.New("blog-rpc", "etc/blog.yaml", run))
}

func run(parent context.Context, c config.Config) error {
	managed := governance.FromContext(parent)
	c.Middlewares.Stat = false
	ctx := svc.NewServiceContext(c)
	managed.SetReadiness(func(probeCtx context.Context) error {
		_, err := healthlogic.NewHealthLogic(probeCtx, ctx).Health(&pb.HealthRequest{})
		return err
	})
	defer ctx.Close()
	s := zrpc.MustNewServer(c.RpcServerConf, func(s *grpc.Server) { pb.RegisterBlogServer(s, server.NewBlogServer(ctx)) })
	defer s.Stop()
	s.AddUnaryInterceptors(managed.Traffic.Unary())
	s.AddStreamInterceptors(managed.Traffic.Stream())
	s.AddUnaryInterceptors(rpcx.UnarySecurity(c.ServiceKey), rpcx.ClientPolicy(ctx.CheckClient))
	s.AddStreamInterceptors(rpcx.StreamSecurity(c.ServiceKey), rpcx.StreamClientPolicy(ctx.CheckClient))
	return managed.Serve(parent, c.ListenOn, s.Start)
}
