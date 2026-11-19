package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type ViewTrendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewViewTrendLogic(ctx context.Context, s *svc.ServiceContext) *ViewTrendLogic {
	return &ViewTrendLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ViewTrendLogic) ViewTrend(req *types.ViewTrendRequest) (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Statistics.GetViewTrend(int(req.Days))
	return ReplyAs[payload.ViewTrendData](value, err)
}
