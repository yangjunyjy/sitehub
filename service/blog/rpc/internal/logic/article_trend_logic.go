package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleTrendLogic(ctx context.Context, s *svc.ServiceContext) *ArticleTrendLogic {
	return &ArticleTrendLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ArticleTrendLogic) ArticleTrend(in *pb.DaysRequest) (*pb.ArticleTrendData, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Statistics.GetArticleTrend(int(in.Days))
	return reply[pb.ArticleTrendData](value, err)
}
