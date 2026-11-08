package svc

import (
	"context"
	"fmt"
	"sitehub/common/rpcx"
	"sitehub/common/runtimeconfig"
	pb "sitehub/service/blog/rpc/pb/blog"
	"sitehub/service/scheduler/cronjob/internal/config"
	logic "sitehub/service/scheduler/cronjob/internal/logic"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct{ Jobs logic.SchedulerService }
type blogJobs struct {
	client  pb.BlogClient
	key     string
	traffic *runtimeconfig.Manager
}

func (b *blogJobs) PublishScheduledArticles(ctx context.Context) (int64, error) {
	release, err := b.traffic.Enter()
	if err != nil {
		return 0, err
	}
	defer release()
	out, err := b.client.PublishScheduled(rpcx.Outgoing(ctx, b.key), &pb.Empty{})
	if err != nil {
		return 0, err
	}
	return out.Count, nil
}
func (b *blogJobs) CleanupArticleAssets(ctx context.Context) error {
	release, err := b.traffic.Enter()
	if err != nil {
		return err
	}
	defer release()
	_, err = b.client.CleanupAssets(rpcx.Outgoing(ctx, b.key), &pb.Empty{})
	return err
}
func NewServiceContext(c config.Config, traffic *runtimeconfig.Manager) (*ServiceContext, error) {
	if len(c.ServiceKey) < 32 {
		return nil, fmt.Errorf("ServiceKey must contain at least 32 characters")
	}
	ports := &blogJobs{pb.NewBlogClient(zrpc.MustNewClient(c.BlogRPC).Conn()), c.ServiceKey, traffic}
	jobs, err := logic.NewWithConfig(ports, ports, c)
	if err != nil {
		return nil, err
	}
	return &ServiceContext{Jobs: jobs}, nil
}
