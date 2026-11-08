package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoleLogic(ctx context.Context, s *svc.ServiceContext) *GetRoleLogic {
	return &GetRoleLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetRoleLogic) GetRole(in *pb.IDRequest) (*pb.EntitySysRole, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:role:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Roles.GetByID(in.Id)
	return reply[pb.EntitySysRole](value, err)
}
