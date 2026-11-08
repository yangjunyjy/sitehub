package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuLogic(ctx context.Context, s *svc.ServiceContext) *GetMenuLogic {
	return &GetMenuLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetMenuLogic) GetMenu(in *pb.IDRequest) (*pb.EntitySysMenu, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:menu:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Menus.GetByID(in.Id)
	return reply[pb.EntitySysMenu](value, err)
}
