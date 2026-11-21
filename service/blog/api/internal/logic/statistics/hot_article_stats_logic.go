package statistics

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type HotArticleStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHotArticleStatsLogic(ctx context.Context, s *svc.ServiceContext) *HotArticleStatsLogic {
	return &HotArticleStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *HotArticleStatsLogic) HotArticleStats(req *types.HotArticleStatsRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Statistics.GetHotArticles(int(req.Limit))
	return logicshared.ReplyAs[payload.StatisticsRows](map[string]any{"items": items}, err)
}
