package logic

import (
	"context"
	"sitehub/service/blog/rpc/internal/svc"
	pb "sitehub/service/blog/rpc/pb/blog"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetArticleStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleStatsLogic(ctx context.Context, s *svc.ServiceContext) *GetArticleStatsLogic {
	return &GetArticleStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetArticleStatsLogic) GetArticleStats(in *pb.Empty) (*pb.ArticleStats, error) {
	actor, err := l.svcCtx.Require(l.ctx, "blog:article:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Statistics.GetArticleStats()
	return reply[pb.ArticleStats](value, err)
}
