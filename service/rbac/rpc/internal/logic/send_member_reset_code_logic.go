package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMemberResetCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMemberResetCodeLogic(ctx context.Context, s *svc.ServiceContext) *SendMemberResetCodeLogic {
	return &SendMemberResetCodeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SendMemberResetCodeLogic) SendMemberResetCode(in *pb.SendCodeRequest) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Members.SendResetCode(l.ctx, in.Email, rpcx.Incoming(l.ctx, "x-client-ip")))
}
