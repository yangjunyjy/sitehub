package systems

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func ListSystems(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.Empty{}, services.RPC.ListSystems)
}

func SaveSystem(ctx context.Context, services *svc.ServiceContext, req *types.SaveSystemRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.SystemEntry{}, services.RPC.SaveSystem)
}

func DeleteSystem(ctx context.Context, services *svc.ServiceContext, req *types.DeleteSystemRequest) (*types.Response, error) {
	return logic.Forward(ctx, services, req, &pb.IDRequest{}, services.RPC.DeleteSystem)
}
