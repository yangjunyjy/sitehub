package icons

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func ListIcons(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.ListIcons)
}

func ValidateIcon(ctx context.Context, services *svc.ServiceContext, req *types.IconInput) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IconRecord{}, services.RPC.ValidateIcon)
}

func CreateIcon(ctx context.Context, services *svc.ServiceContext, req *types.IconInput) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IconRecord{}, services.RPC.CreateIcon)
}

func UpdateIcon(ctx context.Context, services *svc.ServiceContext, req *types.UpdateIconRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IconRecord{}, services.RPC.UpdateIcon)
}

func DeleteIcon(ctx context.Context, services *svc.ServiceContext, req *types.DeleteIconRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.DeleteIcon)
}
