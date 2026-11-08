package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
)

type CaptchaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCaptchaLogic(ctx context.Context, s *svc.ServiceContext) *CaptchaLogic {
	return &CaptchaLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *CaptchaLogic) Captcha(in *pb.Empty) (*pb.SliderCaptchaResult, error) {
	value, err := l.svcCtx.Core.Captcha.Generate(l.ctx)
	return reply[pb.SliderCaptchaResult](value, err)
}
