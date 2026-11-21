package statistics

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type ArticleTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleTrendLogic(ctx context.Context, s *svc.ServiceContext) *ArticleTrendLogic {
	return &ArticleTrendLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ArticleTrendLogic) ArticleTrend(req *types.ArticleTrendRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Statistics.GetArticleTrend(int(req.Days))
	return logicshared.ReplyAs[payload.ArticleTrendData](value, err)
}
