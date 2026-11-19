package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"sitehub/service/product/api/internal/svc"
	"sitehub/service/product/api/internal/types"
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
	pool, err := l.svcCtx.DB.DB()
	if err != nil {
		return nil, err
	}
	return response(map[string]string{"service": "product-api", "status": "ok"}, pool.PingContext(l.ctx))
}
