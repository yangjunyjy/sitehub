package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type ViewTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewViewTrendLogic(ctx context.Context, s *svc.ServiceContext) *ViewTrendLogic {
	return &ViewTrendLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ViewTrendLogic) ViewTrend(in *pb.DaysRequest) (*pb.ViewTrendData, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Statistics.GetViewTrend(int(in.Days))
	return reply[pb.ViewTrendData](value, err)
}
