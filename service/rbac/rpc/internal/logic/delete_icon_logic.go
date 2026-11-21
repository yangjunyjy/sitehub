package logic

import (
	"context"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteIconLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteIconLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteIconLogic {
	return &DeleteIconLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteIconLogic) DeleteIcon(in *rbac.IDRequest) (*rbac.Empty, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:icon:delete", "admin", true); err != nil {
		return nil, err
	}
	return done(l.svcCtx.Core.Icons.Delete(l.ctx, in.Id))
}
