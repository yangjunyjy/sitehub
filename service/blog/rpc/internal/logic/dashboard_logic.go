package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDashboardLogic(ctx context.Context, s *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DashboardLogic) Dashboard(in *pb.Empty) (*pb.DashboardStats, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Statistics.GetDashboardStats()
	return reply[pb.DashboardStats](value, err)
}
