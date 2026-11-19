package logic

import (
	"context"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
)

type MenuListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuListLogic(ctx context.Context, s *svc.ServiceContext) *MenuListLogic {
	return &MenuListLogic{ctx, s}
}
func (l *MenuListLogic) MenuList(req *types.SystemRequest) (*types.Response, error) {
	return NewMenuTreeLogic(l.ctx, l.svcCtx).MenuTree(req)
}
