package logic

import (
	"context"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSystemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSystemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSystemLogic {
	return &DeleteSystemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSystemLogic) DeleteSystem(in *rbac.IDRequest) (*rbac.Empty, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:system:delete", "admin", true); err != nil {
		return nil, err
	}
	return done(l.svcCtx.Core.Systems.Delete(l.ctx, in.Id))
}
