package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"sitehub/common/rpcx"
	"sitehub/service/blog/api/internal/svc"
	"sitehub/service/blog/api/internal/types"
)

type HealthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHealthLogic(ctx context.Context, s *svc.ServiceContext) *HealthLogic {
	return &HealthLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *HealthLogic) Health() (*types.Response, error) {

	if err := l.svcCtx.Health(l.ctx); err != nil {
		return nil, rpcx.Error(err)
	}
	return reply(map[string]string{"service": "blog-api", "status": "ok"}, nil)
}
