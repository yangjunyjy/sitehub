package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, s *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *LoginLogic) Login(in *pb.SysLoginRequest) (*pb.AdminSession, error) {
	return l.svcCtx.Login(l.ctx, in)
}
