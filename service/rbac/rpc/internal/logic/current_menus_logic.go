package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrentMenusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCurrentMenusLogic(ctx context.Context, s *svc.ServiceContext) *CurrentMenusLogic {
	return &CurrentMenusLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CurrentMenusLogic) CurrentMenus(in *pb.Empty) (*pb.MenuPage, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Permissions.GetUserMenus(actor.Id)
	return reply[pb.MenuPage](map[string]any{"items": value}, err)
}
