package statistics

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	logicshared "sitehub/service/blog/api/internal/logic"
	"sitehub/service/blog/api/internal/payload"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type DashboardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDashboardLogic(ctx context.Context, s *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DashboardLogic) Dashboard() (*types.Response, error) {

	_, err := l.svcCtx.Require(l.ctx, "blog:statistics:list", "admin", false)
	if err != nil {
		return nil, err
	}
	value, err := l.svcCtx.Core.Statistics.GetDashboardStats()
	return logicshared.ReplyAs[payload.DashboardStats](value, err)
}
