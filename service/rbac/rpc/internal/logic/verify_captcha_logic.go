package logic

import (
	"context"
	"sitehub/common/rpcx"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyCaptchaLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyCaptchaLogic(ctx context.Context, s *svc.ServiceContext) *VerifyCaptchaLogic {
	return &VerifyCaptchaLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *VerifyCaptchaLogic) VerifyCaptcha(in *pb.CaptchaVerifyRequest) (*pb.BoolReply, error) {
	ok, err := l.svcCtx.Core.Captcha.Verify(l.ctx, in.CaptchaKey, int(in.SliderX), int(in.Duration), int(in.MoveCount))
	if err != nil {
		return nil, rpcx.Error(err)
	}
	if ok {
		err = l.svcCtx.Cache.Set(l.ctx, "native:captcha:"+in.CaptchaKey, "1", 5*time.Minute).Err()
	}
	return &pb.BoolReply{Value: ok}, rpcx.Error(err)
}
