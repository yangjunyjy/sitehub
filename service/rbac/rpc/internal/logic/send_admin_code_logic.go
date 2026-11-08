package logic

import (
	"context"
	"sitehub/service/rbac/rpc/internal/svc"
	pb "sitehub/service/rbac/rpc/pb/rbac"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SendAdminCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendAdminCodeLogic(ctx context.Context, s *svc.ServiceContext) *SendAdminCodeLogic {
	return &SendAdminCodeLogic{ctx, s, logx.WithContext(ctx)}
}
func (l *SendAdminCodeLogic) SendAdminCode(in *pb.SendCodeRequest) (*pb.Empty, error) {
	proof, err := l.svcCtx.Cache.GetDel(l.ctx, "native:captcha:"+in.CaptchaKey).Result()
	if err != nil || proof != "1" {
		return nil, status.Error(codes.InvalidArgument, "captcha verification required")
	}
	return done(l.svcCtx.Core.Users.GenerateEmailCode(l.ctx, in.Email, ""))
}
