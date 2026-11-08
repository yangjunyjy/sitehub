package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type CategoryStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryStatsLogic(ctx context.Context, s *svc.ServiceContext) *CategoryStatsLogic {
	return &CategoryStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CategoryStatsLogic) CategoryStats(in *pb.LimitRequest) (*pb.Records, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	items, err := l.svcCtx.Core.Statistics.GetCategoryStats()
	return reply[pb.Records](map[string]any{"items": items}, err)
}
