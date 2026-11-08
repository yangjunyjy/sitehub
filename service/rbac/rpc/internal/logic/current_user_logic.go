package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrentUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCurrentUserLogic(ctx context.Context, s *svc.ServiceContext) *CurrentUserLogic {
	return &CurrentUserLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CurrentUserLogic) CurrentUser(in *pb.Empty) (*pb.EntitySysUser, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	value, err := l.svcCtx.Core.Users.GetUserInfo(actor.Id)
	return reply[pb.EntitySysUser](value, err)
}
