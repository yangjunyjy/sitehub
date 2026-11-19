package main

import (
	"context"
	"fmt"
	"sitehub/common/apiresponse"
	"sitehub/common/command"
	"sitehub/common/governance"
	"sitehub/common/rpcx"
	"sitehub/service/product/api/internal/config"
	"sitehub/service/product/api/internal/handler"
	healthlogic "sitehub/service/product/api/internal/logic"
	"sitehub/service/product/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func main() {
	command.Execute(command.New("product-api", "etc/product-api.yaml", run))
}

func run(parent context.Context, c config.Config) error {
	managed := governance.FromContext(parent)
	c.Middlewares.Log = false
	s := rest.MustNewServer(c.RestConf)
	defer s.Stop()
	managed.AttachHTTP(s)
	s.Use(rpcx.HTTPContextWithProxies(c.TrustedProxyCIDRs))
	apiresponse.Install()
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		return err
	}
	defer ctx.Close()
	s.Use(ctx.ClientPolicy)
	managed.SetReadiness(func(probeCtx context.Context) error {
		_, err := healthlogic.NewHealthLogic(probeCtx, ctx).Health()
		return err
	})
	handler.RegisterHandlers(s, ctx)
	handler.RegisterInternalHandlers(s, ctx)
	return managed.Serve(parent, fmt.Sprintf("%s:%d", c.Host, c.Port), s.Start)
}
