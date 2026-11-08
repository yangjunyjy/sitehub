package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, s *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *LogoutLogic) Logout(in *pb.Empty) (*pb.Empty, error) {
	actor, err := l.svcCtx.Require(l.ctx, "", "admin", false)
	if err != nil {
		return nil, err
	}
	_ = actor
	return done(l.svcCtx.Core.Sessions.Delete(rpcx.Incoming(l.ctx, "authorization")))
}
