package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutMemberLogic(ctx context.Context, s *svc.ServiceContext) *LogoutMemberLogic {
	return &LogoutMemberLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *LogoutMemberLogic) LogoutMember(in *pb.Empty) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Members.Logout(l.ctx, rpcx.Incoming(l.ctx, "authorization")))
}
