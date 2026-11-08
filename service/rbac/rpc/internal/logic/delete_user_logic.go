package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, s *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *DeleteUserLogic) DeleteUser(in *pb.IDRequest) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "system:user:delete", "admin", true)
	if err != nil {
		return nil, err
	}
	_ = actor

	err = l.svcCtx.Core.Users.DeleteUser(in.Id, actor.Id)
	return done(err)
}
