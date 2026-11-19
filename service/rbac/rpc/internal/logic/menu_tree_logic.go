package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuTreeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMenuTreeLogic(ctx context.Context, s *svc.ServiceContext) *MenuTreeLogic {
	return &MenuTreeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *MenuTreeLogic) MenuTree(in *pb.SystemRequest) (*pb.MenuPage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:menu:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Systems.MenuTree(l.ctx, in.System, false)
	return reply[pb.MenuPage](map[string]any{"items": value}, err)
}
