package captcha

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func Captcha(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.Captcha)
}

func VerifyCaptcha(ctx context.Context, services *svc.ServiceContext, req *types.VerifyCaptchaRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.CaptchaVerifyRequest{}, services.RPC.VerifyCaptcha)
}
