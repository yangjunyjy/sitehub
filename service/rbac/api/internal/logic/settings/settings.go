package settings

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func GetEmailSettings(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.GetEmailSettings)
}

func SaveEmailSettings(ctx context.Context, services *svc.ServiceContext, req *types.SaveEmailSettingsRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.UpdateEmailSettingsRequest{}, services.RPC.SaveEmailSettings)
}
