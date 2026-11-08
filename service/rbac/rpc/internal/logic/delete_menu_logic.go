package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMenuLogic(ctx context.Context, s *svc.ServiceContext) *DeleteMenuLogic {
	return &DeleteMenuLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteMenuLogic) DeleteMenu(in *pb.IDRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:menu:delete", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Menus.Delete(in.Id)
	return done(err)
}
