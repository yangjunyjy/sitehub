package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRoleLogic(ctx context.Context, s *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteRoleLogic) DeleteRole(in *pb.IDRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:role:delete", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Roles.Delete(in.Id)
	return done(err)
}
