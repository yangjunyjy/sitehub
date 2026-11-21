package statistics

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type GetArticleStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetArticleStatsLogic(ctx context.Context, s *svc.ServiceContext) *GetArticleStatsLogic {
	return &GetArticleStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetArticleStatsLogic) GetArticleStats() (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Statistics.GetArticleStats()
	return logicshared.ReplyAs[payload.ArticleStats](value, err)
}
