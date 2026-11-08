package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type TOTPStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTOTPStatusLogic(ctx context.Context, s *svc.ServiceContext) *TOTPStatusLogic {
	return &TOTPStatusLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *TOTPStatusLogic) TOTPStatus(in *pb.Empty) (*pb.TOTPInfo, error) {
	return &pb.TOTPInfo{Enabled: l.svcCtx.Config.TOTP.Enabled, TimeStep: int64(l.svcCtx.Config.TOTP.TimeStep), WindowSize: int64(l.svcCtx.Config.TOTP.WindowSize)}, nil
}
