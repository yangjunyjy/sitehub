package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type TagStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTagStatsLogic(ctx context.Context, s *svc.ServiceContext) *TagStatsLogic {
	return &TagStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *TagStatsLogic) TagStats(in *pb.LimitRequest) (*pb.Records, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	items, err := l.svcCtx.Core.Statistics.GetTagStats()
	return reply[pb.Records](map[string]any{"items": items}, err)
}
