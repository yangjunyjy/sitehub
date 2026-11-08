package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRoleMenusLogic(ctx context.Context, s *svc.ServiceContext) *RoleMenusLogic {
	return &RoleMenusLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *RoleMenusLogic) RoleMenus(in *pb.IDRequest) (*pb.MenuPage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:role:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Roles.GetRoleMenus(in.Id)
	return reply[pb.MenuPage](map[string]any{"items": value}, err)
}
