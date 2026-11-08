package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMemberCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMemberCodeLogic(ctx context.Context, s *svc.ServiceContext) *SendMemberCodeLogic {
	return &SendMemberCodeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SendMemberCodeLogic) SendMemberCode(in *pb.SendCodeRequest) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Members.SendRegistrationCode(l.ctx, in.Email, rpcx.Incoming(l.ctx, "x-client-ip")))
}
