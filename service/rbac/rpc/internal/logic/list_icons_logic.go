package logic

import (
	"context"

	"sitehub/service/rbac/rpc/internal/svc"
	"sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListIconsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListIconsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListIconsLogic {
	return &ListIconsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListIconsLogic) ListIcons(in *rbac.Empty) (*rbac.IconPage, error) {
	if _, err := l.svcCtx.Require(l.ctx, "system:icon:list", "admin", false); err != nil {
		return nil, err
	}
	items, err := l.svcCtx.Core.Icons.List(l.ctx)
	return reply[rbac.IconPage](map[string]any{"items": items}, err)
}
