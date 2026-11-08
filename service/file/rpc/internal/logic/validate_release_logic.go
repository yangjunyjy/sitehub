package logic

import (
	"context"
	"sitehub/service/file/rpc/internal/svc"
	pb "sitehub/service/file/rpc/pb/file"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateReleaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateReleaseLogic(ctx context.Context, s *svc.ServiceContext) *ValidateReleaseLogic {
	return &ValidateReleaseLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *ValidateReleaseLogic) ValidateRelease(in *pb.LocationRequest) (*pb.Empty, error) {
	return done(l.svcCtx.Core.Settings.ValidateRelease(in.Location))
}
