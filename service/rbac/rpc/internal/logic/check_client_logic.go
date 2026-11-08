package logic

import (
	"context"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckClientLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckClientLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckClientLogic {
	return &CheckClientLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckClientLogic) CheckClient(in *rbac.Empty) (*rbac.Empty, error) {
	if err := l.svcCtx.CheckClient(l.ctx); err != nil {
		return nil, err
	}
	return &rbac.Empty{}, nil
}
