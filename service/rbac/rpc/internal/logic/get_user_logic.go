package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, s *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *GetUserLogic) GetUser(in *pb.IDRequest) (*pb.EntitySysUser, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:user:list", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Users.GetUserInfo(in.Id)
	return reply[pb.EntitySysUser](value, err)
}
