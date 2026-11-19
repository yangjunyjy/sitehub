package logic

import (
	"context"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSystemsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSystemsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSystemsLogic {
	return &ListSystemsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSystemsLogic) ListSystems(in *rbac.Empty) (*rbac.SystemPage, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:system:list", "admin", false); err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Systems.List(l.ctx)
	return reply[rbac.SystemPage](map[string]any{"items": items}, err)
}
