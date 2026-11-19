package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type CategoryStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCategoryStatsLogic(ctx context.Context, s *svc.ServiceContext) *CategoryStatsLogic {
	return &CategoryStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CategoryStatsLogic) CategoryStats(req *types.CategoryStatsRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Statistics.GetCategoryStats()
	return ReplyAs[payload.StatisticsRows](map[string]any{"items": items}, err)
}
