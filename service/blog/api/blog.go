package main

import (
	"context"
	"fmt"
	"sitehub/common/apiresponse"
	"sitehub/common/command"
	"sitehub/common/governance"
	"sitehub/common/rpcx"
	"sitehub/common/runtimeconfig"
	"sitehub/service/blog/api/internal/config"
	"sitehub/service/blog/api/internal/handler"
	healthlogic "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/jobs"

	"github.com/zeromicro/go-zero/rest"
)

func main() {
	command.Execute(command.New("blog-api", "etc/blog-api.yaml", run))
}

func run(parent context.Context, c config.Config) error {
	managed := governance.FromContext(parent)
	c.Middlewares.Log = false
	s := rest.MustNewServer(c.RestConf)
	defer s.Stop()
	managed.AttachHTTP(s)
	s.Use(rpcx.HTTPContextWithProxies(c.TrustedProxyCIDRs))
	apiresponse.Install()
	ctx, err := svc.New(c)
	if err != nil {
		return err
	}
	defer ctx.Close()
	s.Use(ctx.ClientPolicy)
	if c.Scheduler.Enabled {
		local := &localJobs{ctx.Core.Articles, ctx.Core.Assets, managed.Traffic}
		scheduler, err := jobs.NewWithConfig(local, local, c.Scheduler)
		if err != nil {
			return err
		}
		scheduler.Start()
		defer scheduler.Stop()
	}
	managed.SetReadiness(func(probeCtx context.Context) error {
		_, err := healthlogic.NewHealthLogic(probeCtx, ctx).Health()
		return err
	})
	registerHandlers(s, ctx)
	return managed.Serve(parent, fmt.Sprintf("%s:%d", c.Host, c.Port), s.Start)
}

func registerHandlers(server *rest.Server, ctx *svc.ServiceContext) {
	handler.RegisterHandlers(server, ctx)
	handler.RegisterStreamHandlers(server, ctx)
	handler.RegisterStorageHTTP(server, ctx)
}

type localJobs struct {
	publisher jobs.Publisher
	cleaner   jobs.Cleaner
	traffic   *runtimeconfig.Manager
}

func (j *localJobs) PublishScheduledArticles(ctx context.Context) (int64, error) {
	release, err := j.traffic.Enter()
	if err != nil {
		return 0, err
	}
	defer release()
	return j.publisher.PublishScheduledArticles(ctx)
}

func (j *localJobs) CleanupArticleAssets(ctx context.Context) error {
	release, err := j.traffic.Enter()
	if err != nil {
		return err
	}
	defer release()
	return j.cleaner.CleanupArticleAssets(ctx)
}
