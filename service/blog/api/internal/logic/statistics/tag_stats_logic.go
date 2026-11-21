package statistics

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type TagStatsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTagStatsLogic(ctx context.Context, s *svc.ServiceContext) *TagStatsLogic {
	return &TagStatsLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *TagStatsLogic) TagStats(req *types.TagStatsRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Statistics.GetTagStats()
	return logicshared.ReplyAs[payload.StatisticsRows](map[string]any{"items": items}, err)
}
