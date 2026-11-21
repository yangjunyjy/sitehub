package auth

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func ConsumeToken(ctx context.Context, services *svc.ServiceContext, req *types.ConsumeTokenRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.ConsumeTokenRequest{Token: req.Token}, services.RPC.ConsumeToken)
}

func SendAdminCode(ctx context.Context, services *svc.ServiceContext, req *types.SendAdminCodeRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.SendCodeRequest{}, services.RPC.SendAdminCode)
}

func CurrentUser(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.CurrentUser)
}

func CurrentPermissions(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.CurrentPermissions)
}

func CurrentMenus(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.CurrentMenus)
}

func OnceToken(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.OnceToken)
}

func TOTPStatus(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.TOTPStatus)
}

func ChangePassword(ctx context.Context, services *svc.ServiceContext, req *types.ChangePasswordRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.UpdatePasswordRequest{}, services.RPC.ChangePassword)
}
