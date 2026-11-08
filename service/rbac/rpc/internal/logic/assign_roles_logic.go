package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRolesLogic(ctx context.Context, s *svc.ServiceContext) *AssignRolesLogic {
	return &AssignRolesLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *AssignRolesLogic) AssignRoles(in *pb.AssignRolesRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:user:edit", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Users.AssignRoles(in.UserId, in.RoleIds)
	return done(err)
}
